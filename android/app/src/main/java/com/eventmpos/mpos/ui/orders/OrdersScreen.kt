package com.eventmpos.mpos.ui.orders

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderState

/** BA order history for the event session, including not-yet-confirmed (Pending) submissions. */
@Composable
fun OrdersScreen(viewModel: OrdersViewModel = hiltViewModel()) {
    val orders by viewModel.orders.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Text("Order history", style = MaterialTheme.typography.titleLarge)
        if (orders.isEmpty()) {
            Text("No orders yet for this event.", color = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.padding(top = 12.dp))
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize().padding(top = 8.dp)) {
                items(orders, key = { it.idempotencyKey }) { OrderRow(it) }
            }
        }
    }
}

@Composable
private fun OrderRow(order: OrderRecord) {
    Card(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
        Row(modifier = Modifier.fillMaxWidth().padding(14.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f)) {
                Text(
                    order.orderNumberDisplay ?: "Order (queued)",
                    fontWeight = FontWeight.SemiBold,
                )
                Text(
                    order.customerName + (order.referenceId?.let { " · $it" } ?: ""),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                order.total?.let { Text("RM ${"%.2f".format(it)}", style = MaterialTheme.typography.bodySmall) }
                order.error?.let { Text(it, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.error) }
            }
            StatusChip(order.state)
        }
    }
}

@Composable
private fun StatusChip(state: OrderState) {
    val (label, color) = when (state) {
        OrderState.PENDING -> "Pending" to MaterialTheme.colorScheme.tertiary
        OrderState.CONFIRMED -> "Confirmed" to MaterialTheme.colorScheme.primary
        OrderState.FAILED -> "Failed" to MaterialTheme.colorScheme.error
    }
    Text(label, color = color, fontWeight = FontWeight.SemiBold, style = MaterialTheme.typography.labelMedium)
}
