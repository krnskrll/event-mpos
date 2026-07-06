package com.eventmpos.mpos.ui.session

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.CalendarMonth
import androidx.compose.material.icons.outlined.Place
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.components.Pill
import com.eventmpos.mpos.ui.components.SectionCard
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun EventSelectionScreen(
    onStarted: () -> Unit,
    onLoggedOut: () -> Unit,
    viewModel: EventSelectionViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().background(Brand.Bg)) {
        // Header
        Column(modifier = Modifier.fillMaxWidth().background(Brand.Navy).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 20.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(modifier = Modifier.size(40.dp).clip(CircleShape).background(Brand.Teal), contentAlignment = Alignment.Center) {
                    Text(initials(viewModel.userName), color = Color.White, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                }
                Spacer(Modifier.size(12.dp))
                Column(Modifier.weight(1f)) {
                    Text("Welcome,", color = Color.White.copy(alpha = 0.6f), fontSize = 13.sp)
                    Text(viewModel.userName, color = Color.White, fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
                Box(
                    modifier = Modifier.clip(RoundedCornerShape(8.dp)).background(Color.White.copy(alpha = 0.12f))
                        .clickable { viewModel.logout(onLoggedOut) }
                        .padding(horizontal = 12.dp, vertical = 8.dp),
                ) { Text("Logout", color = Color.White.copy(alpha = 0.85f), fontSize = 13.sp) }
            }
        }

        Text(
            "YOUR ASSIGNED EVENTS",
            color = Brand.TextFaint, fontSize = 13.sp, fontWeight = FontWeight.Medium, letterSpacing = 1.sp,
            modifier = Modifier.padding(start = 20.dp, top = 20.dp, bottom = 8.dp),
        )

        when {
            state.loading -> Box(Modifier.fillMaxSize(), Alignment.Center) { CircularProgressIndicator(color = Brand.Navy) }
            state.error != null -> Column(Modifier.fillMaxSize().padding(20.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text(state.error!!, color = Brand.Danger)
                BrandButton("Retry", viewModel::load, modifier = Modifier.padding(top = 12.dp))
            }
            state.events.isEmpty() -> Text("No events assigned to you yet.", color = Brand.TextMuted, modifier = Modifier.padding(20.dp))
            else -> LazyColumn(contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)) {
                items(state.events, key = { it.id }) { ev ->
                    EventCard(ev, onStart = { viewModel.start(ev); onStarted() })
                    Spacer(Modifier.height(12.dp))
                }
            }
        }
    }
}

@Composable
private fun EventCard(event: AssignedEvent, onStart: () -> Unit) {
    val active = event.status == EventStatus.ACTIVE
    SectionCard {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            if (active) Pill("● ACTIVE", Brand.Ready, Brand.ReadyBg) else Pill(event.status.name, Brand.Packing, Brand.PackingBg)
        }
        Text(event.name, color = Brand.Heading, fontSize = 18.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(top = 10.dp, bottom = 6.dp))
        MetaRow(Icons.Outlined.CalendarMonth, "${event.startDate ?: "—"} → ${event.endDate ?: "—"}")
        Spacer(Modifier.size(2.dp))
        MetaRow(Icons.Outlined.Place, event.code)
        Spacer(Modifier.size(14.dp))
        BrandButton(
            text = if (active) "Start Session" else "Not active yet",
            onClick = onStart,
            enabled = active,
            color = Brand.Teal,
            modifier = Modifier.fillMaxWidth(),
        )
    }
}

@Composable
private fun MetaRow(icon: androidx.compose.ui.graphics.vector.ImageVector, text: String) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(icon, contentDescription = null, tint = Brand.TextFaint, modifier = Modifier.size(15.dp))
        Spacer(Modifier.size(6.dp))
        Text(text, color = Brand.TextMuted, fontSize = 13.sp)
    }
}

private fun initials(name: String) =
    name.split(" ").mapNotNull { it.firstOrNull()?.uppercase() }.take(2).joinToString("")
