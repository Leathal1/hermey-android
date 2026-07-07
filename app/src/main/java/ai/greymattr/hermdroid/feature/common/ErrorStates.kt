package ai.greymattr.hermdroid.feature.common

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CloudOff
import androidx.compose.material.icons.filled.Error
import androidx.compose.material.icons.filled.SignalWifiOff
import androidx.compose.material.icons.filled.SyncProblem
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Button
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp

sealed class ErrorType(val title: String, val message: String)
object NetworkDown : ErrorType("Network unavailable", "Check your connection and try again.")
object AuthExpired : ErrorType("Session expired", "Sign in again to continue.")
object StreamDropped : ErrorType("Stream interrupted", "The response stream was dropped. Retry?")
object ServerUnreachable : ErrorType("Server unreachable", "Hermes Agent did not respond. Try later.")

@Composable
fun ErrorState(
    error: ErrorType,
    modifier: Modifier = Modifier,
    onRetry: (() -> Unit)? = null,
    onSignIn: (() -> Unit)? = null
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        val icon = when (error) {
            NetworkDown -> Icons.Default.SignalWifiOff
            AuthExpired -> Icons.Default.Warning
            StreamDropped -> Icons.Default.SyncProblem
            ServerUnreachable -> Icons.Default.CloudOff
            else -> Icons.Default.Error
        }
        Icon(
            imageVector = icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.error,
            modifier = Modifier.height(64.dp)
        )
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = error.title,
            style = MaterialTheme.typography.titleLarge,
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = error.message,
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(modifier = Modifier.height(24.dp))
        onRetry?.let {
            Button(onClick = it) {
                Text("Retry")
            }
        }
        if (error AuthExpired) {
            onSignIn?.let {
                Spacer(modifier = Modifier.height(8.dp))
                Button(onClick = it) {
                    Text("Sign in")
                }
            }
        }
    }
}
