package com.eventmpos.mpos.ui.receipt

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

/**
 * Shows the stored 58mm receipt (monospace, exactly as printed) with a Reprint button. Reprint
 * re-sends to the printer with no new order and no network — and surfaces printer failures + retry.
 */
@Composable
fun ReceiptScreen(viewModel: ReceiptViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Text("Receipt", style = MaterialTheme.typography.titleLarge)
        Spacer(Modifier.height(12.dp))

        when {
            state.loading -> CircularProgressIndicator()
            state.lines.isEmpty() -> Text(state.message ?: "No receipt.", color = MaterialTheme.colorScheme.error)
            else -> {
                Card(modifier = Modifier.fillMaxWidth().weight(1f)) {
                    Text(
                        text = state.lines.joinToString("\n"),
                        fontFamily = FontFamily.Monospace,
                        fontSize = 12.sp,
                        textAlign = TextAlign.Start,
                        modifier = Modifier.fillMaxWidth().verticalScroll(rememberScrollState()).padding(12.dp),
                    )
                }
                Spacer(Modifier.height(12.dp))
                Button(
                    onClick = viewModel::reprint,
                    enabled = !state.printing,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(if (state.printing) "Printing…" else "Reprint receipt")
                }
            }
        }

        state.message?.let {
            if (state.lines.isNotEmpty()) {
                Spacer(Modifier.height(8.dp))
                Text(it, color = if (state.error) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary)
            }
        }
    }
}
