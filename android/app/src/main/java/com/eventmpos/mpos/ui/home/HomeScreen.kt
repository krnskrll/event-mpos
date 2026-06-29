package com.eventmpos.mpos.ui.home

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.auth.AssignedEvent

/**
 * Post-login home for a BA. Loads assigned events, auto-selects when there's only one, and downloads
 * the event snapshot before any selling. Enforces the lifecycle: only ACTIVE events can be synced.
 */
@Composable
fun HomeScreen(
    onLoggedOut: () -> Unit,
    onBrowseProducts: (String) -> Unit,
    onIdentifyCustomer: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(state.loggedOut) {
        if (state.loggedOut) onLoggedOut()
    }

    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        when {
            state.loadingEvents -> {
                CircularProgressIndicator()
                Spacer(Modifier.height(12.dp))
                Text("Loading your events…")
            }

            state.loadError != null -> {
                Text(state.loadError!!, color = MaterialTheme.colorScheme.error)
                Spacer(Modifier.height(12.dp))
                OutlinedButton(onClick = viewModel::loadEvents) { Text("Retry") }
            }

            state.events.isEmpty() -> {
                Text("No event is assigned to your account yet.")
                Text(
                    "Ask an admin to assign you to an event.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }

            state.selected == null -> EventPicker(state.events, onPick = viewModel::select)

            else -> SelectedEvent(
                event = state.selected!!,
                sync = state.sync,
                showBack = state.events.size > 1,
                onRetry = viewModel::retrySync,
                onBack = viewModel::backToList,
                onBrowseProducts = onBrowseProducts,
                onIdentifyCustomer = onIdentifyCustomer,
            )
        }

        Spacer(Modifier.height(28.dp))
        Button(onClick = viewModel::logout, modifier = Modifier.fillMaxWidth()) { Text("Log out") }
    }
}

@Composable
private fun EventPicker(events: List<AssignedEvent>, onPick: (AssignedEvent) -> Unit) {
    Text("Select your event", style = MaterialTheme.typography.headlineSmall)
    Spacer(Modifier.height(12.dp))
    LazyColumn(modifier = Modifier.fillMaxWidth()) {
        items(events) { event ->
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 6.dp)
                    .clickable { onPick(event) },
            ) {
                Column(Modifier.padding(16.dp)) {
                    Text(event.name, fontWeight = FontWeight.SemiBold)
                    Text(
                        "${event.code} · ${event.status}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            }
        }
    }
}

@Composable
private fun SelectedEvent(
    event: AssignedEvent,
    sync: SyncUi,
    showBack: Boolean,
    onRetry: () -> Unit,
    onBack: () -> Unit,
    onBrowseProducts: (String) -> Unit,
    onIdentifyCustomer: () -> Unit,
) {
    Text("Your event", style = MaterialTheme.typography.headlineSmall)
    Card(modifier = Modifier.fillMaxWidth().padding(top = 16.dp)) {
        Column(Modifier.padding(20.dp)) {
            Text(event.name, style = MaterialTheme.typography.titleLarge)
            Text(
                "${event.code} · ${event.status}",
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.padding(top = 4.dp),
            )
            if (event.startDate != null || event.endDate != null) {
                Text(
                    "${event.startDate ?: "—"} → ${event.endDate ?: "—"}",
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 8.dp),
                )
            }
            Spacer(Modifier.height(16.dp))
            SyncStatus(sync, onRetry)
            if (sync is SyncUi.Ready) {
                Button(
                    onClick = { onBrowseProducts(event.id) },
                    modifier = Modifier.fillMaxWidth().padding(top = 16.dp),
                ) {
                    Text("Browse products")
                }
                OutlinedButton(
                    onClick = onIdentifyCustomer,
                    modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                ) {
                    Text("Identify customer")
                }
            }
        }
    }
    if (showBack) {
        TextButton(onClick = onBack, modifier = Modifier.padding(top = 8.dp)) {
            Text("← Choose a different event")
        }
    }
}

@Composable
private fun SyncStatus(sync: SyncUi, onRetry: () -> Unit) {
    when (sync) {
        SyncUi.Idle -> {}

        SyncUi.Syncing -> Row(verticalAlignment = Alignment.CenterVertically) {
            CircularProgressIndicator(strokeWidth = 2.dp, modifier = Modifier.height(18.dp))
            Spacer(Modifier.height(0.dp))
            Text("  Downloading event data…")
        }

        is SyncUi.Ready -> Text(
            if (sync.unchanged) "✓ Up to date (v${sync.version}). Ready to sell."
            else "✓ Synced (v${sync.version}). Ready to sell.",
            color = MaterialTheme.colorScheme.primary,
            fontWeight = FontWeight.SemiBold,
        )

        is SyncUi.Blocked -> Text(sync.message, color = MaterialTheme.colorScheme.error)

        is SyncUi.Failed -> Column {
            Text(sync.message, color = MaterialTheme.colorScheme.error)
            Spacer(Modifier.height(8.dp))
            OutlinedButton(onClick = onRetry) { Text("Retry sync") }
        }
    }
}
