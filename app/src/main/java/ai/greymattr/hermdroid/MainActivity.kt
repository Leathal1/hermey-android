package ai.greymattr.hermdroid

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import ai.greymattr.hermdroid.ui.theme.HermdroidTheme
import ai.greymattr.hermdroid.feature.inbox.InboxScreen
import ai.greymattr.hermdroid.feature.chat.chatScreen
import ai.greymattr.hermdroid.feature.chat.navigateToChat
import ai.greymattr.hermdroid.feature.settings.SettingsScreen

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        val splashScreen = installSplashScreen()
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        splashScreen.setKeepOnScreenCondition { false }

        setContent {
            HermdroidTheme {
                HermdroidApp()
            }
        }
    }
}

@Composable
fun HermdroidApp() {
    val navController = rememberNavController()
    Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = "inbox",
            modifier = Modifier.padding(innerPadding)
        ) {
            composable("inbox") {
                InboxScreen(
                    onSessionClick = { session ->
                        navController.navigateToChat(session.id, session.title)
                    },
                    onSettingsClick = { navController.navigate("settings") }
                )
            }
            composable("settings") {
                SettingsScreen(onBack = { navController.popBackStack() })
            }
            chatScreen(onBack = { navController.popBackStack() })
        }
    }
}
