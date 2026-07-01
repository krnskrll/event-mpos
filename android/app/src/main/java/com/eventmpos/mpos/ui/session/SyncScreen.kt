package com.eventmpos.mpos.ui.session

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun SyncScreen(
    onSynced: () -> Unit,
    viewModel: SyncViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val progress by animateFloatAsState(targetValue = state.progress, label = "sync")

    LaunchedEffect(state.done) { if (state.done) { kotlinx.coroutines.delay(700); onSynced() } }

    Column(
        modifier = Modifier.fillMaxSize().background(Color.White).padding(horizontal = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text("SETTING UP EVENT", color = Brand.TextFaint, fontSize = 13.sp, fontWeight = FontWeight.Medium, letterSpacing = 1.sp)
        Text(viewModel.eventName, color = Brand.Heading, fontSize = 20.sp, fontWeight = FontWeight.Bold, textAlign = TextAlign.Center, modifier = Modifier.padding(top = 6.dp, bottom = 32.dp))

        Box(
            modifier = Modifier.size(88.dp).clip(CircleShape).background(if (state.done) Brand.TealSoft else Brand.NavActiveBg),
            contentAlignment = Alignment.Center,
        ) {
            Text(if (state.done) "✓" else "${(progress * 100).toInt()}%", color = if (state.done) Brand.Teal else Brand.Navy, fontSize = if (state.done) 40.sp else 18.sp, fontWeight = FontWeight.Bold)
        }

        Text(
            if (state.error != null) state.error!!
            else if (state.done) "All done! Starting session…"
            else viewModel.steps.getOrElse(state.stepIndex) { "" },
            color = if (state.error != null) Brand.Danger else Brand.TextMuted,
            fontSize = 15.sp, textAlign = TextAlign.Center,
            modifier = Modifier.padding(vertical = 24.dp),
        )

        LinearProgressIndicator(
            progress = { progress },
            color = Brand.Teal,
            trackColor = Brand.CardBorder,
            modifier = Modifier.fillMaxWidth().height(6.dp).clip(RoundedCornerShape(999.dp)),
        )

        Spacer(Modifier.height(28.dp))
        viewModel.steps.forEachIndexed { i, step ->
            val doneStep = i < state.stepIndex || state.done
            val current = i == state.stepIndex && !state.done
            Row(Modifier.fillMaxWidth().padding(vertical = 4.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier.size(20.dp).clip(CircleShape).background(if (doneStep) Brand.Teal else if (current) Brand.NavActiveBg else Brand.CardBorder),
                    contentAlignment = Alignment.Center,
                ) { if (doneStep) Text("✓", color = Color.White, fontSize = 11.sp) }
                Spacer(Modifier.size(12.dp))
                Text(step, color = if (doneStep) Brand.Teal else if (current) Brand.Heading else Brand.TextGhost, fontSize = 13.sp, fontWeight = if (current) FontWeight.SemiBold else FontWeight.Normal)
            }
        }

        if (state.error != null) {
            BrandButton("Retry", viewModel::run, modifier = Modifier.fillMaxWidth().padding(top = 24.dp))
        }
    }
}
