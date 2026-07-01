package com.eventmpos.mpos.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderState
import com.eventmpos.mpos.ui.components.Pill
import com.eventmpos.mpos.ui.components.SectionCard
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun OrdersScreen(
    onOpenReceipt: (String) -> Unit,
    viewModel: OrdersViewModel = hiltViewModel(),
) {
    val orders by viewModel.orders.collectAsStateWithLifecycle()

    Column(Modifier.fillMaxSize().background(Brand.Bg)) {
        Column(Modifier.fillMaxWidth().background(Brand.Navy).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 16.dp)) {
            Text("Order history", color = Color.White, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Text("${orders.size} orders this session", color = Color.White.copy(alpha = 0.6f), fontSize = 13.sp)
        }
        if (orders.isEmpty()) {
            Box(Modifier.fillMaxSize(), Alignment.Center) { Text("No orders yet for this event.", color = Brand.TextMuted) }
        } else {
            LazyColumn(contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)) {
                items(orders, key = { it.idempotencyKey }) { o ->
                    OrderRow(o, onOpenReceipt); Spacer(Modifier.size(10.dp))
                }
            }
        }
    }
}

@Composable
private fun OrderRow(order: OrderRecord, onOpen: (String) -> Unit) {
    val confirmed = order.state == OrderState.CONFIRMED
    SectionCard(modifier = Modifier.then(if (confirmed) Modifier.clickable { onOpen(order.idempotencyKey) } else Modifier)) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f)) {
                Text(order.orderNumberDisplay ?: "Order (queued)", color = Brand.Heading, fontSize = 16.sp, fontWeight = FontWeight.Bold)
                Text(order.customerName + (order.referenceId?.let { " · $it" } ?: ""), color = Brand.TextMuted, fontSize = 12.sp)
                order.total?.let { Text("RM ${"%.2f".format(it)}", color = Brand.Text, fontSize = 13.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 2.dp)) }
                order.error?.let { Text(it, color = Brand.Danger, fontSize = 12.sp) }
            }
            Column(horizontalAlignment = Alignment.End) {
                when (order.state) {
                    OrderState.PENDING -> Pill("Pending", Brand.Packing, Brand.PackingBg)
                    OrderState.CONFIRMED -> Pill("Confirmed", Brand.Ready, Brand.ReadyBg)
                    OrderState.FAILED -> Pill("Failed", Brand.Danger, Brand.CollectedBg)
                }
                if (confirmed) Text("Receipt ›", color = Brand.Teal, fontSize = 12.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 8.dp))
            }
        }
    }
}
