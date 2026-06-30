package com.eventmpos.mpos.ui.cart

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
import androidx.compose.material3.Divider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

/**
 * The cart — entirely local. Adjust quantities, see the auto-applied promotion preview, and confirm.
 * Confirm requires an identified customer (EPIC-04). On confirm the order is queued + submitted via
 * WorkManager; the BA is taken to order history where it shows Pending → Confirmed.
 */
@Composable
fun CartScreen(
    onConfirmed: () -> Unit,
    onIdentifyCustomer: () -> Unit,
    viewModel: CartViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Text("Cart", style = MaterialTheme.typography.titleLarge)

        if (state.lines.isEmpty()) {
            Text("Cart is empty. Add products from the catalogue.", color = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.padding(top = 12.dp))
        } else {
            LazyColumn(modifier = Modifier.weight(1f)) {
                items(state.lines, key = { it.product.id }) { line ->
                    Card(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth().padding(12.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Column(Modifier.weight(1f)) {
                                Text(line.product.name, fontWeight = FontWeight.SemiBold)
                                Text("RM ${"%.2f".format(line.product.price)} · line RM ${"%.2f".format(line.lineTotal)}",
                                    style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                            OutlinedButton(onClick = { viewModel.decrement(line.product.id) }) { Text("−") }
                            Text("  ${line.qty}  ", fontWeight = FontWeight.SemiBold)
                            OutlinedButton(onClick = { viewModel.increment(line.product.id) }) { Text("+") }
                        }
                    }
                }
            }

            Spacer(Modifier.height(8.dp))
            // customer
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Text(
                    state.customer?.let { "Customer: ${it.name}${it.tier?.let { t -> " ($t)" } ?: ""}" } ?: "No customer selected",
                    color = if (state.customer == null) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.onSurface,
                    style = MaterialTheme.typography.bodyMedium,
                )
                OutlinedButton(onClick = onIdentifyCustomer) { Text(if (state.customer == null) "Identify" else "Change") }
            }

            Spacer(Modifier.height(8.dp))
            Divider()
            TotalRow("Subtotal", state.subtotal)
            if (state.discount > 0.0) TotalRow("Discount (${state.promotionName ?: "promo"})", -state.discount, MaterialTheme.colorScheme.primary)
            TotalRow("Total", state.total, bold = true)

            Spacer(Modifier.height(12.dp))
            Button(
                onClick = { viewModel.confirm(onConfirmed) },
                enabled = viewModel.canConfirm && !state.submitting,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(
                    when {
                        state.submitting -> "Submitting…"
                        state.customer == null -> "Identify a customer to confirm"
                        else -> "Confirm order"
                    },
                )
            }
        }
    }
}

@Composable
private fun TotalRow(label: String, amount: Double, color: androidx.compose.ui.graphics.Color = MaterialTheme.colorScheme.onSurface, bold: Boolean = false) {
    Row(modifier = Modifier.fillMaxWidth().padding(vertical = 2.dp), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = color, fontWeight = if (bold) FontWeight.Bold else FontWeight.Normal)
        Text("RM ${"%.2f".format(amount)}", color = color, fontWeight = if (bold) FontWeight.Bold else FontWeight.Normal)
    }
}
