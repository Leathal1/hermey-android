package ai.greymattr.hermdroid.feature.skills

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import ai.greymattr.hermdroid.feature.common.EmptyState
import ai.greymattr.hermdroid.feature.common.NoSkills

@Composable
fun SkillsScreen(modifier: Modifier = Modifier) {
    EmptyState(NoSkills, modifier)
}
