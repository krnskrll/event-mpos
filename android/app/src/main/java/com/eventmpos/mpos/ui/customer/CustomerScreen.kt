package com.eventmpos.mpos.ui.customer

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.member.CustomerSource

/**
 * Identify the customer for the order: scan/type a QR (resolved from the offline Room cache, then an
 * online fallback), or type a name manually if neither works. The selected customer is what an
 * order will be linked to in EPIC-07 (an order can't be submitted without one).
 */
@Composable
fun CustomerScreen(viewModel: CustomerViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var qr by remember { mutableStateOf("") }
    var manualName by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize().padding(20.dp)) {
        Text("Identify customer", style = MaterialTheme.typography.titleLarge)
        Spacer(Modifier.height(12.dp))

        OutlinedTextField(
            value = qr,
            onValueChange = { qr = it },
            label = { Text("Scan or type customer QR (e.g. MBR-00001)") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
        )
        Button(
            onClick = { viewModel.identifyByQr(qr) },
            enabled = !state.identifying,
            modifier = Modifier.padding(top = 8.dp),
        ) {
            Text(if (state.identifying) "Identifying…" else "Identify")
        }

        if (state.identifying) {
            Spacer(Modifier.height(12.dp))
            CircularProgressIndicator()
        }

        state.selected?.let { c ->
            Card(modifier = Modifier.fillMaxWidth().padding(top = 16.dp)) {
                Column(Modifier.padding(16.dp)) {
                    Text("Customer for this order", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Text(c.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                    Text(
                        buildString {
                            append(c.tier ?: "—")
                            append(" · ")
                            append(
                                when (c.source) {
                                    CustomerSource.LOCAL_CACHE -> "matched offline"
                                    CustomerSource.ONLINE_LOOKUP -> "matched online"
                                    CustomerSource.MANUAL -> "manually entered"
                                },
                            )
                        },
                        style = MaterialTheme.typography.bodySmall,
                    )
                    OutlinedButton(onClick = viewModel::clear, modifier = Modifier.padding(top = 8.dp)) {
                        Text("Change")
                    }
                }
            }
        }

        // QR didn't resolve anywhere -> manual name fallback
        if (state.notFoundQr != null && state.selected == null) {
            Spacer(Modifier.height(16.dp))
            Text(
                "No customer found for \"${state.notFoundQr}\" (offline cache or online). Enter a name to proceed:",
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
            OutlinedTextField(
                value = manualName,
                onValueChange = { manualName = it },
                label = { Text("Customer name") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
            )
            Button(onClick = { viewModel.setManual(manualName) }, modifier = Modifier.padding(top = 8.dp)) {
                Text("Use this name")
            }
        }
    }
}
