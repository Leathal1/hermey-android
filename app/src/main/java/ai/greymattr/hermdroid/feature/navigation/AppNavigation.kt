package ai.greymattr.hermdroid.feature.navigation

import ai.greymattr.hermdroid.feature.chat.ChatScreen
import ai.greymattr.hermdroid.feature.newsession.NewSessionScreen
import ai.greymattr.hermdroid.feature.sessions.SessionsScreen
import ai.greymattr.hermdroid.feature.settings.SettingsScreen
import ai.greymattr.hermdroid.feature.skills.SkillsScreen
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Chat
import androidx.compose.material.icons.filled.Dashboard
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController

const val ChatRoute = "chat_home"
const val SessionsRoute = "sessions"
const val WorkspaceRoute = "workspace"
const val SettingsRoute = "settings"
const val NewSessionRoute = "new_session"
const val ActiveChatRoute = "active_chat/{sessionId}"
const val ActiveChatSessionIdArg = "sessionId"

sealed class TopLevelRoute(val name: String, val route: String, val icon: ImageVector) {
    data object Chat : TopLevelRoute("Chat", ChatRoute, Icons.AutoMirrored.Filled.Chat)
    data object Sessions : TopLevelRoute("Sessions", SessionsRoute, Icons.Default.Dashboard)
    data object Workspace : TopLevelRoute("Workspace", WorkspaceRoute, Icons.Default.Folder)
    data object Settings : TopLevelRoute("Settings", SettingsRoute, Icons.Default.Settings)
}

@Composable
fun HermdroidApp() {
    val navController = rememberNavController()
    val currentBackStack by navController.currentBackStackEntryAsState()
    val currentRoute = currentBackStack?.destination?.route
    val showBottomBar = listOf(TopLevelRoute.Chat, TopLevelRoute.Sessions, TopLevelRoute.Workspace, TopLevelRoute.Settings).any {
        currentRoute?.startsWith(it.route) == true
    }

    Scaffold(
        bottomBar = {
            if (showBottomBar) {
                BottomNavBar(navController)
            }
        }
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = ChatRoute,
            modifier = Modifier.padding(innerPadding)
        ) {
            topLevelGraph(navController)
        }
    }
}

private fun NavGraphBuilder.topLevelGraph(navController: NavHostController) {
    composable(ChatRoute) {
        Text("Chat home — start a new session from the Sessions tab")
    }
    composable(SessionsRoute) {
        SessionsScreen(
            onSessionClick = { session ->
                navController.navigate("active_chat/${session.id}?title=${session.title}")
            },
            onNewSessionClick = { navController.navigate(NewSessionRoute) },
            onSettingsClick = { navController.navigate(SettingsRoute) }
        )
    }
    composable(WorkspaceRoute) {
        SkillsScreen()
    }
    composable(SettingsRoute) {
        SettingsScreen(onBack = { navController.popBackStack() })
    }
    composable(NewSessionRoute) {
        NewSessionScreen(
            onSessionCreated = { id, title ->
                navController.popBackStack()
                navController.navigate("active_chat/$id?title=$title")
            },
            onBack = { navController.popBackStack() }
        )
    }
    composable(
        route = "active_chat/{$ActiveChatSessionIdArg}?title={title}",
        arguments = listOf(
            androidx.navigation.navArgument(ActiveChatSessionIdArg) { type = androidx.navigation.NavType.StringType },
            androidx.navigation.navArgument("title") {
                type = androidx.navigation.NavType.StringType
                defaultValue = ""
            }
        )
    ) { backStackEntry ->
        val sessionId = backStackEntry.arguments?.getString(ActiveChatSessionIdArg) ?: ""
        val title = backStackEntry.arguments?.getString("title") ?: ""
        ChatScreen(
            sessionId = sessionId,
            sessionTitle = title,
            onBack = { navController.popBackStack() }
        )
    }
}

@Composable
private fun BottomNavBar(navController: NavHostController) {
    val currentBackStack by navController.currentBackStackEntryAsState()
    val currentRoute = currentBackStack?.destination?.route
    NavigationBar {
        listOf(TopLevelRoute.Chat, TopLevelRoute.Sessions, TopLevelRoute.Workspace, TopLevelRoute.Settings).forEach { routeObj ->
            val selected = currentRoute?.startsWith(routeObj.route) == true
            NavigationBarItem(
                icon = { Icon(routeObj.icon, contentDescription = routeObj.name) },
                label = { Text(routeObj.name) },
                selected = selected,
                onClick = {
                    navController.navigate(routeObj.route) {
                        popUpTo(navController.graph.startDestinationId) {
                            saveState = true
                        }
                        launchSingleTop = true
                        restoreState = true
                    }
                }
            )
        }
    }
}
