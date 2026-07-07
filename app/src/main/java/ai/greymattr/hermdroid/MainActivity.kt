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
import ai.greymattr.hermdroid.ui.theme.HermdroidTheme
import ai.greymattr.hermdroid.feature.inbox.InboxScreen

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        val splashScreen = installSplashScreen()
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        splashScreen.setKeepOnScreenCondition { false }

        setContent {
            HermdroidTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    InboxScreen(modifier = Modifier.padding(innerPadding))
                }
            }
        }
    }
}
