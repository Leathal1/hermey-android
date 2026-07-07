package ai.greymattr.hermdroid.feature.sessions

import ai.greymattr.hermdroid.data.repository.ModelOption
import ai.greymattr.hermdroid.data.repository.ProfileOption
import ai.greymattr.hermdroid.data.repository.SessionRepository
import ai.greymattr.hermdroid.data.repository.SessionResult
import ai.greymattr.hermdroid.data.repository.UiProject
import ai.greymattr.hermdroid.data.repository.UiSession
import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch

sealed class SessionsUiState {
    object Loading : SessionsUiState()
    data class Success(
        val sessions: List<UiSession>,
        val query: String,
        val sort: SessionSort,
        val showArchived: Boolean
    ) : SessionsUiState()
    data class Error(val offline: Boolean, val message: String) : SessionsUiState()
}

enum class SessionSort {
    Recents, Name, MessageCount
}

sealed class ActionResult {
    object Idle : ActionResult()
    data class Done(val message: String) : ActionResult()
    data class Error(val message: String) : ActionResult()
}

class SessionsViewModel(application: Application) : AndroidViewModel(application) {

    private val repo = SessionRepository(application)

    private val _all = MutableStateFlow<List<UiSession>>(emptyList())
    private val _query = MutableStateFlow("")
    private val _sort = MutableStateFlow(SessionSort.Recents)
    private val _showArchived = MutableStateFlow(false)
    private val _actionResult = MutableStateFlow<ActionResult>(ActionResult.Idle)
    val actionResult: StateFlow<ActionResult> = _actionResult

    private val _projects = MutableStateFlow<List<UiProject>>(emptyList())
    val projects: StateFlow<List<UiProject>> = _projects

    private val _models = MutableStateFlow<List<ModelOption>>(emptyList())
    private val _profiles = MutableStateFlow<List<ProfileOption>>(emptyList())

    val uiState: StateFlow<SessionsUiState> = combine(
        _all, _query, _sort, _showArchived
    ) { all, query, sort, archived ->
        var list = all.filter { archived || !it.archived }
        if (query.isNotBlank()) {
            val q = query.lowercase()
            list = list.filter {
                it.title.lowercase().contains(q) ||
                        it.model.lowercase().contains(q) ||
                        it.profile.lowercase().contains(q)
            }
        }
        list = when (sort) {
            SessionSort.Recents -> list.sortedByDescending { it.lastMessageAtMs }
            SessionSort.Name -> list.sortedBy { it.title.lowercase() }
            SessionSort.MessageCount -> list.sortedByDescending { it.messageCount }
        }
        SessionsUiState.Success(list, query, sort, archived)
    }.let { it as StateFlow<SessionsUiState> }

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            when (val r = repo.sessions()) {
                is SessionResult.Success -> {
                    _all.value = r.value
                    _showArchived.value = false
                }
                is SessionResult.Error -> _all.value = emptyList()
            }
        }
        viewModelScope.launch {
            when (val r = repo.projects()) {
                is SessionResult.Success -> _projects.value = r.value
                is SessionResult.Error -> _projects.value = emptyList()
            }
        }
        viewModelScope.launch {
            when (val r = repo.models()) {
                is SessionResult.Success -> _models.value = r.value
                is SessionResult.Error -> _models.value = emptyList()
            }
        }
        viewModelScope.launch {
            when (val r = repo.profiles()) {
                is SessionResult.Success -> _profiles.value = r.value
                is SessionResult.Error -> _profiles.value = emptyList()
            }
        }
    }

    fun setQuery(query: String) {
        _query.value = query
    }

    fun setSort(sort: SessionSort) {
        _sort.value = sort
    }

    fun toggleArchived() {
        _showArchived.value = !_showArchived.value
    }

    fun rename(sessionId: String, title: String) {
        viewModelScope.launch {
            when (repo.rename(sessionId, title)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.map { if (it.id == sessionId) it.copy(title = title) else it }
                    _actionResult.value = ActionResult.Done("Renamed")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Rename failed")
            }
        }
    }

    fun delete(sessionId: String) {
        viewModelScope.launch {
            when (repo.delete(sessionId)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.filter { it.id != sessionId }
                    _actionResult.value = ActionResult.Done("Deleted")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Delete failed")
            }
        }
    }

    fun pin(sessionId: String, pinned: Boolean) {
        viewModelScope.launch {
            when (repo.pin(sessionId, pinned)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.map { if (it.id == sessionId) it.copy(pinned = pinned) else it }
                    _actionResult.value = ActionResult.Done(if (pinned) "Pinned" else "Unpinned")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Pin failed")
            }
        }
    }

    fun archive(sessionId: String, archived: Boolean) {
        viewModelScope.launch {
            when (repo.archive(sessionId, archived)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.map { if (it.id == sessionId) it.copy(archived = archived) else it }
                    _actionResult.value = ActionResult.Done(if (archived) "Archived" else "Restored")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Archive failed")
            }
        }
    }

    fun branch(sessionId: String, title: String, onBranched: (String) -> Unit) {
        viewModelScope.launch {
            when (val r = repo.branch(sessionId, title)) {
                is SessionResult.Success -> {
                    _actionResult.value = ActionResult.Done("Branched")
                    onBranched(r.value)
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Branch failed")
            }
        }
    }

    fun truncate(sessionId: String) {
        viewModelScope.launch {
            when (repo.truncate(sessionId, "", 0)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.map { if (it.id == sessionId) it.copy(messageCount = 0) else it }
                    _actionResult.value = ActionResult.Done("Truncated")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Truncate failed")
            }
        }
    }

    fun moveToProject(sessionId: String, projectId: String) {
        viewModelScope.launch {
            when (repo.moveToProject(sessionId, projectId)) {
                is SessionResult.Success -> {
                    _all.value = _all.value.map { if (it.id == sessionId) it.copy(projectId = projectId) else it }
                    _actionResult.value = ActionResult.Done("Moved")
                }
                is SessionResult.Error -> _actionResult.value = ActionResult.Error("Move failed")
            }
        }
    }

    fun consumeActionResult() {
        _actionResult.value = ActionResult.Idle
    }
}
