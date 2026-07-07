package ai.greymattr.hermdroid.feature.chat

import androidx.compose.runtime.Composable
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument

const val ChatRoute = "chat/{sessionId}"
const val ChatSessionIdArg = "sessionId"
const val ChatSessionTitleArg = "title"

fun NavController.navigateToChat(sessionId: String, title: String = "") {
    navigate("chat/$sessionId?title=$title")
}

fun NavGraphBuilder.chatScreen(onBack: () -> Unit) {
    composable(
        route = "chat/{$ChatSessionIdArg}?title={$ChatSessionTitleArg}",
        arguments = listOf(
            navArgument(ChatSessionIdArg) { type = NavType.StringType },
            navArgument(ChatSessionTitleArg) {
                type = NavType.StringType
                defaultValue = ""
            }
        )
    ) { backStackEntry ->
        val sessionId = backStackEntry.arguments?.getString(ChatSessionIdArg) ?: ""
        val title = backStackEntry.arguments?.getString(ChatSessionTitleArg) ?: ""
        ChatScreen(
            sessionId = sessionId,
            sessionTitle = title,
            onBack = onBack
        )
    }
}
