package ai.greymattr.hermdroid.feature.chat

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel

@Composable
fun MessageList(
    messages: List<UiMessage>,
    onToggleReasoning: (String) -> Unit,
    onToggleTool: (String, String) -> Unit,
    modifier: Modifier = Modifier,
    viewModel: ChatViewModel = viewModel()
) {
    val listState = rememberLazyListState()
    val scrollToBottom by viewModel.scrollToBottom.collectAsState()

    LaunchedEffect(scrollToBottom) {
        if (scrollToBottom && messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size - 1)
            viewModel.onScrollToBottomConsumed()
        }
    }

    LazyColumn(
        state = listState,
        modifier = modifier.fillMaxSize()
    ) {
        items(messages, key = { it.id }) { msg ->
            Box(modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)) {
                Text(
                    text = "${msg.role}: ${msg.displayContent}",
                    color = when (msg.role) {
                        MessageRole.User -> MaterialTheme.colorScheme.primary
                        MessageRole.Assistant -> MaterialTheme.colorScheme.onSurface
                        MessageRole.System -> MaterialTheme.colorScheme.tertiary
                    }
                )
            }
        }
    }
}
