package ai.greymattr.hermdroid.feature.inbox

import ai.greymattr.hermdroid.core.cache.CachedSession
import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

sealed class InboxUiState {
    object Loading : InboxUiState()
    data class Success(val sessions: List<CachedSession>) : InboxUiState()
    data class Error(val offline: Boolean) : InboxUiState()
}

class InboxViewModel(application: Application) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow<InboxUiState>(InboxUiState.Loading)
    val uiState: StateFlow<InboxUiState> = _uiState

    private val repo by lazy { ai.greymattr.hermdroid.data.repository.OfflineRepository(application) }

    init {
        refresh()
    }

    fun refresh() {
        _uiState.value = InboxUiState.Loading
        viewModelScope.launch {
            // In a full implementation this would also try a network fetch and fall back to cache.
            val sessions = repo.sessions()
            _uiState.value = InboxUiState.Success(sessions)
        }
    }
}
