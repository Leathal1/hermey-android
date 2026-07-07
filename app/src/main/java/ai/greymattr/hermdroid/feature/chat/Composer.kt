package ai.greymattr.hermdroid.feature.chat

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel

@Composable
fun Composer(
    viewModel: ChatViewModel = viewModel(),
    modifier: Modifier = Modifier
) {
    val input by viewModel.inputText.collectAsState()
    Column(modifier = modifier.padding(8.dp)) {
        OutlinedTextField(
            value = input,
            onValueChange = viewModel::onInputChange,
            placeholder = { Text("Message") },
            modifier = Modifier.fillMaxWidth(),
            trailingIcon = {
                IconButton(onClick = viewModel::send) {
                    Icon(Icons.AutoMirrored.Filled.Send, contentDescription = "Send")
                }
            }
        )
    }
}

@Composable
private fun Text(text: String) {
    androidx.compose.material3.Text(text)
}
