package ai.greymattr.hermdroid.feature.newsession

import ai.greymattr.hermdroid.data.local.SecureSettings
import ai.greymattr.hermdroid.data.repository.ModelOption
import ai.greymattr.hermdroid.data.repository.ProfileOption
import ai.greymattr.hermdroid.data.repository.SessionRepository
import ai.greymattr.hermdroid.data.repository.SessionResult
import ai.greymattr.hermdroid.data.repository.UiSession
import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

sealed class NewSessionUiState {
    object Loading : NewSessionUiState()
    data class Ready(
        val models: List<ModelOption>,
        val profiles: List<ProfileOption>,
        val selectedModel: String,
        val selectedProfile: String,
        val workspace: String,
        val title: String
    ) : NewSessionUiState()
    data class Created(val session: UiSession) : NewSessionUiState()
    data class Error(val message: String) : NewSessionUiState()
}

class NewSessionViewModel(application: Application) : AndroidViewModel(application) {

    private val repo = SessionRepository(application)
    private val _uiState = MutableStateFlow<NewSessionUiState>(NewSessionUiState.Loading)
    val uiState: StateFlow<NewSessionUiState> = _uiState

    init {
        load()
    }

    private fun load() {
        viewModelScope.launch {
            val models = when (val r = repo.models()) {
                is SessionResult.Success -> r.value
                is SessionResult.Error -> emptyList()
            }
            val profiles = when (val r = repo.profiles()) {
                is SessionResult.Success -> r.value
                is SessionResult.Error -> emptyList()
            }
            _uiState.value = NewSessionUiState.Ready(
                models = models,
                profiles = profiles,
                selectedModel = models.firstOrNull { it.id == SecureSettings.defaultModel }?.id ?: models.firstOrNull()?.id ?: SecureSettings.defaultModel,
                selectedProfile = profiles.firstOrNull { it.isActive }?.id ?: profiles.firstOrNull()?.id ?: "",
                workspace = "",
                title = ""
            )
        }
    }

    fun selectModel(model: String) {
        val state = _uiState.value as? NewSessionUiState.Ready ?: return
        _uiState.value = state.copy(selectedModel = model)
    }

    fun selectProfile(profile: String) {
        val state = _uiState.value as? NewSessionUiState.Ready ?: return
        _uiState.value = state.copy(selectedProfile = profile)
    }

    fun setWorkspace(workspace: String) {
        val state = _uiState.value as? NewSessionUiState.Ready ?: return
        _uiState.value = state.copy(workspace = workspace)
    }

    fun setTitle(title: String) {
        val state = _uiState.value as? NewSessionUiState.Ready ?: return
        _uiState.value = state.copy(title = title)
    }

    fun create() {
        val state = _uiState.value as? NewSessionUiState.Ready ?: return
        viewModelScope.launch {
            when (val r = repo.newSession(state.workspace, state.selectedModel, state.selectedProfile, state.title)) {
                is SessionResult.Success -> _uiState.value = NewSessionUiState.Created(r.value)
                is SessionResult.Error -> _uiState.value = NewSessionUiState.Error(r.message)
            }
        }
    }
}
