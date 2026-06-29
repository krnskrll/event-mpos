package com.eventmpos.mpos.ui.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

/**
 * Post-login home. Shows ONLY the BA's assigned event (EPIC-01 requirement). Later epics load the
 * event snapshot (products/stock/members/promotions) from here and open the selling flow.
 */
@Composable
fun HomeScreen(
    onLoggedOut: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(state.loggedOut) {
        if (state.loggedOut) onLoggedOut()
    }

    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Your event", style = MaterialTheme.typography.headlineSmall)

        when {
            state.loading -> CircularProgressIndicator(modifier = Modifier.padding(top = 24.dp))

            state.event != null -> {
                val event = state.event!!
                Card(modifier = Modifier.fillMaxWidth().padding(top = 16.dp)) {
                    Column(modifier = Modifier.padding(20.dp)) {
                        Text(event.name, style = MaterialTheme.typography.titleLarge)
                        Text(
                            "Code: ${event.code}",
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
                    }
                }
            }

            state.error != null -> {
                Text(
                    state.error!!,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(top = 16.dp),
                )
                OutlinedButton(onClick = viewModel::loadEvent, modifier = Modifier.padding(top = 12.dp)) {
                    Text("Retry")
                }
            }
        }

        Button(
            onClick = viewModel::logout,
            modifier = Modifier.fillMaxWidth().padding(top = 32.dp),
        ) {
            Text("Log out")
        }
    }
}
