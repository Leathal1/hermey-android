package ai.greymattr.hermdroid.feature.sessions

import ai.greymattr.hermdroid.data.repository.UiProject
import ai.greymattr.hermdroid.data.repository.UiSession
import ai.greymattr.hermdroid.feature.common.EmptyState
import ai.greymattr.hermdroid.feature.common.ErrorState
import ai.greymattr.hermdroid.feature.common.NetworkDown
import ai.greymattr.hermdroid.feature.common.NoSessions
import ai.greymattr.hermdroid.feature.common.ServerUnreachable
import ai.greymattr.hermdroid.feature.common.SkeletonList
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Archive
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material.icons.filled.PushPin
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Sort
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SessionsScreen(
    onSessionClick: (UiSession) -> Unit,
    onNewSessionClick: () -> Unit,
    onSettingsClick: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: SessionsViewModel = viewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val actionResult by viewModel.actionResult.collectAsState()
    val projects by viewModel.projects.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(actionResult) {
        when (val r = actionResult) {
            is ActionResult.Done -> {
                snackbarHostState.showSnackbar(r.message)
                viewModel.consumeActionResult()
            }
            is ActionResult.Error -> {
                snackbarHostState.showSnackbar(r.message)
                viewModel.consumeActionResult()
            }
            else -> {}
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Sessions") },
                actions = {
                    IconButton(onClick = { viewModel.toggleArchived() }) {
                        Icon(Icons.Default.Archive, contentDescription = "Toggle archived")
                    }
                    IconButton(onClick = onSettingsClick) {
                        Icon(Icons.Default.Settings, contentDescription = "Settings")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = onNewSessionClick) {
                Icon(Icons.Default.Add, contentDescription = "New session")
            }
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        Column(
            modifier = modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            SearchField(
                query = (uiState as? SessionsUiState.Success)?.query ?: "",
                onQueryChange = viewModel::setQuery,
                onSortChange = viewModel::setSort,
                modifier = Modifier.fillMaxWidth()
            )
            when (val state = uiState) {
                is SessionsUiState.Loading -> SkeletonList()
                is SessionsUiState.Error -> ErrorState(
                    error = if (state.offline) NetworkDown else ServerUnreachable,
                    onRetry = { viewModel.refresh() }
                )
                is SessionsUiState.Success -> {
                    if (state.sessions.isEmpty()) {
                        EmptyState(NoSessions)
                    } else {
                        LazyColumn {
                            items(state.sessions, key = { it.id }) { session ->
                                SessionRow(
                                    session = session,
                                    projects = projects,
                                    onClick = { onSessionClick(session) },
                                    onRename = { viewModel.rename(session.id, it) },
                                    onDelete = { viewModel.delete(session.id) },
                                    onPin = { viewModel.pin(session.id, !session.pinned) },
                                    onArchive = { viewModel.archive(session.id, !session.archived) },
                                    onBranch = { viewModel.branch(session.id, it) { id -> onSessionClick(session.copy(id = id, title = it)) } },
                                    onTruncate = { viewModel.truncate(session.id) },
                                    onMove = { viewModel.moveToProject(session.id, it) }
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun SearchField(
    query: String,
    onQueryChange: (String) -> Unit,
    onSortChange: (SessionSort) -> Unit,
    modifier: Modifier = Modifier
) {
    var expanded by remember { mutableStateOf(false) }
    Row(
        modifier = modifier.padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        OutlinedTextField(
            value = query,
            onValueChange = onQueryChange,
            modifier = Modifier.weight(1f),
            placeholder = { Text("Search sessions") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            singleLine = true
        )
        Spacer(modifier = Modifier.width(8.dp))
        IconButton(onClick = { expanded = true }) {
            Icon(Icons.Default.Sort, contentDescription = "Sort")
        }
        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            SessionSort.values().forEach { sort ->
                DropdownMenuItem(
                    text = { Text(sort.name) },
                    onClick = {
                        onSortChange(sort)
                        expanded = false
                    }
                )
            }
        }
    }
}

@Composable
private fun SessionRow(
    session: UiSession,
    projects: List<UiProject>,
    onClick: () -> Unit,
    onRename: (String) -> Unit,
    onDelete: () -> Unit,
    onPin: () -> Unit,
    onArchive: () -> Unit,
    onBranch: (String) -> Unit,
    onTruncate: () -> Unit,
    onMove: (String) -> Unit
) {
    var menuExpanded by remember { mutableStateOf(false) }
    var dialog by rememberSaveable { mutableStateOf<SessionDialog?>(null) }

    ListItem(
        headlineContent = { Text(session.title) },
        supportingContent = {
            Column {
                Text(
                    "${session.messageCount} messages • ${session.model.ifBlank { "default" }}",
                    style = MaterialTheme.typography.bodyMedium
                )
                if (session.pinned || session.archived) {
                    Row {
                        if (session.pinned) {
                            Text("Pinned", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary)
                        }
                        if (session.archived) {
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Archived", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.tertiary)
                        }
                    }
                }
            }
        },
        trailingContent = {
            IconButton(onClick = { menuExpanded = true }) {
                Icon(Icons.Default.MoreVert, contentDescription = "Actions")
            }
            DropdownMenu(expanded = menuExpanded, onDismissRequest = { menuExpanded = false }) {
                DropdownMenuItem(
                    text = { Text(if (session.pinned) "Unpin" else "Pin") },
                    leadingIcon = { Icon(Icons.Default.PushPin, contentDescription = null) },
                    onClick = { menuExpanded = false; onPin() }
                )
                DropdownMenuItem(
                    text = { Text("Rename") },
                    leadingIcon = { Icon(Icons.Default.Edit, contentDescription = null) },
                    onClick = { menuExpanded = false; dialog = SessionDialog.Rename }
                )
                DropdownMenuItem(
                    text = { Text(if (session.archived) "Restore" else "Archive") },
                    leadingIcon = { Icon(Icons.Default.Archive, contentDescription = null) },
                    onClick = { menuExpanded = false; onArchive() }
                )
                DropdownMenuItem(
                    text = { Text("Branch") },
                    leadingIcon = { Icon(Icons.AutoMirrored.Filled.List, contentDescription = null) },
                    onClick = { menuExpanded = false; dialog = SessionDialog.Branch }
                )
                DropdownMenuItem(
                    text = { Text("Truncate") },
                    leadingIcon = { Icon(Icons.Default.Delete, contentDescription = null) },
                    onClick = { menuExpanded = false; dialog = SessionDialog.Truncate }
                )
                DropdownMenuItem(
                    text = { Text("Move to project") },
                    leadingIcon = { Icon(Icons.Default.Archive, contentDescription = null) },
                    onClick = { menuExpanded = false; dialog = SessionDialog.Move }
                )
                DropdownMenuItem(
                    text = { Text("Delete") },
                    leadingIcon = { Icon(Icons.Default.Delete, contentDescription = null) },
                    onClick = { menuExpanded = false; dialog = SessionDialog.Delete }
                )
            }
        },
        modifier = Modifier.clickable { onClick() }
    )

    when (dialog) {
        is SessionDialog.Rename -> TextInputDialog(
            title = "Rename session",
            initial = session.title,
            onConfirm = { onRename(it); dialog = null },
            onDismiss = { dialog = null }
        )
        is SessionDialog.Branch -> TextInputDialog(
            title = "Branch session",
            initial = "${session.title} (branch)",
            onConfirm = { onBranch(it); dialog = null },
            onDismiss = { dialog = null }
        )
        is SessionDialog.Truncate -> ConfirmDialog(
            title = "Truncate session?",
            text = "This will remove all messages after the first one.",
            onConfirm = { onTruncate(); dialog = null },
            onDismiss = { dialog = null }
        )
        is SessionDialog.Delete -> ConfirmDialog(
            title = "Delete session?",
            text = "This cannot be undone.",
            onConfirm = { onDelete(); dialog = null },
            onDismiss = { dialog = null }
        )
        is SessionDialog.Move -> ProjectPickerDialog(
            projects = projects,
            onConfirm = { onMove(it); dialog = null },
            onDismiss = { dialog = null }
        )
        null -> {}
    }
}

private sealed class SessionDialog {
    object Rename : SessionDialog()
    object Branch : SessionDialog()
    object Truncate : SessionDialog()
    object Delete : SessionDialog()
    object Move : SessionDialog()
}

@Composable
private fun TextInputDialog(
    title: String,
    initial: String,
    onConfirm: (String) -> Unit,
    onDismiss: () -> Unit
) {
    var text by rememberSaveable { mutableStateOf(initial) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title) },
        text = {
            OutlinedTextField(
                value = text,
                onValueChange = { text = it },
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )
        },
        confirmButton = {
            TextButton(onClick = { onConfirm(text) }, enabled = text.isNotBlank()) {
                Text("Save")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Cancel") }
        }
    )
}

@Composable
private fun ConfirmDialog(
    title: String,
    text: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title) },
        text = { Text(text) },
        confirmButton = {
            TextButton(onClick = onConfirm) { Text("Confirm") }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Cancel") }
        }
    )
}

@Composable
private fun ProjectPickerDialog(
    projects: List<UiProject>,
    onConfirm: (String) -> Unit,
    onDismiss: () -> Unit
) {
    var selected by rememberSaveable { mutableStateOf<String?>(null) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Move to project") },
        text = {
            LazyColumn {
                items(projects, key = { it.id }) { project ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { selected = project.id }
                            .padding(vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(project.name, modifier = Modifier.weight(1f))
                        if (selected == project.id) {
                            Text("✓", color = MaterialTheme.colorScheme.primary)
                        }
                    }
                }
            }
        },
        confirmButton = {
            TextButton(
                onClick = { selected?.let(onConfirm) },
                enabled = selected != null
            ) { Text("Move") }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Cancel") }
        }
    )
}
