package ai.greymattr.hermdroid.feature.tasks

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import ai.greymattr.hermdroid.feature.common.EmptyState
import ai.greymattr.hermdroid.feature.common.NoTasks

@Composable
fun TasksScreen(modifier: Modifier = Modifier) {
    EmptyState(NoTasks, modifier)
}
