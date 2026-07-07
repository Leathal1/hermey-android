package ai.greymattr.hermdroid.feature.inbox

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.lifecycle.viewmodel.compose.viewModel
import ai.greymattr.hermdroid.feature.common.EmptyState
import ai.greymattr.hermdroid.feature.common.ErrorState
import ai.greymattr.hermdroid.feature.common.SkeletonList
import ai.greymattr.hermdroid.feature.common.NetworkDown
import ai.greymattr.hermdroid.feature.common.NoSessions
import ai.greymattr.hermdroid.feature.common.ServerUnreachable

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun InboxScreen(
    modifier: Modifier = Modifier,
    viewModel: InboxViewModel = viewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Hermdroid") },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Warning, contentDescription = "Refresh")
                    }
                    IconButton(onClick = { /* navigate to settings */ }) {
                        Icon(Icons.Default.Settings, contentDescription = "Settings")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when (val state = uiState) {
                is InboxUiState.Loading -> SkeletonList()
                is InboxUiState.Error -> ErrorState(
                    error = if (state.offline) NetworkDown else ServerUnreachable,
                    onRetry = { viewModel.refresh() }
                )
                is InboxUiState.Success -> {
                    if (state.sessions.isEmpty()) {
                        EmptyState(NoSessions)
                    } else {
                        LazyColumn {
                            items(state.sessions, key = { it.id }) { session ->
                                ListItem(
                                    headlineContent = { Text(session.title) },
                                    supportingContent = {
                                        Text(
                                            "${session.messageCount} messages",
                                            style = MaterialTheme.typography.bodyMedium
                                        )
                                    }
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}
