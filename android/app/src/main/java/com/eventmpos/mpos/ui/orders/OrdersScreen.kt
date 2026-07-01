package com.eventmpos.mpos.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Assignment
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderState
import com.eventmpos.mpos.ui.theme.Brand

/** BA-side statuses (the local submission lifecycle) shown as underline filter tabs. */
private val FILTERS = listOf<Pair<String, OrderState?>>(
    "All" to null,
    "Confirmed" to OrderState.CONFIRMED,
    "Pending" to OrderState.PENDING,
    "Failed" to OrderState.FAILED,
)

@Composable
fun OrdersScreen(
    onOpenReceipt: (String) -> Unit,
    viewModel: OrdersViewModel = hiltViewModel(),
) {
    val orders by viewModel.orders.collectAsStateWithLifecycle()
    var filter by remember { mutableStateOf<OrderState?>(null) }
    val filtered = orders.filter { filter == null || it.state == filter }

    Column(Modifier.fillMaxSize().background(Brand.Bg)) {
        // White header
        Row(
            Modifier.fillMaxWidth().background(Color.White).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(Modifier.weight(1f)) {
                Text("Session summary", color = Brand.TextFaint, fontSize = 12.sp)
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("Today's Orders", color = Brand.Heading, fontSize = 18.sp, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.size(8.dp))
                    Box(Modifier.clip(RoundedCornerShape(999.dp)).background(Brand.NavActiveBg).padding(horizontal = 10.dp, vertical = 3.dp)) {
                        Text("${orders.size}", color = Brand.Navy, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
        // Underline filter tabs
        Row(Modifier.fillMaxWidth().background(Color.White).padding(horizontal = 16.dp)) {
            FILTERS.forEach { (label, state) ->
                val active = filter == state
                val count = if (state == null) orders.size else orders.count { it.state == state }
                Column(
                    Modifier.clickable { filter = state }.padding(horizontal = 14.dp, vertical = 12.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(label, color = if (active) Brand.Navy else Brand.TextFaint, fontSize = 13.sp, fontWeight = if (active) FontWeight.Bold else FontWeight.Normal)
                        if (count > 0) {
                            Spacer(Modifier.size(5.dp))
                            Box(Modifier.clip(RoundedCornerShape(999.dp)).background(if (active) Brand.Navy else Brand.CardBorder).padding(horizontal = 6.dp, vertical = 1.dp)) {
                                Text("$count", color = if (active) Color.White else Brand.TextFaint, fontSize = 10.sp, fontWeight = FontWeight.SemiBold)
                            }
                        }
                    }
                    Spacer(Modifier.size(6.dp))
                    Box(Modifier.size(width = 24.dp, height = 2.dp).background(if (active) Brand.Navy else Color.Transparent))
                }
            }
        }

        if (filtered.isEmpty()) {
            Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.Center) {
                Box(Modifier.size(72.dp).clip(CircleShape).background(Brand.CardBorder), Alignment.Center) {
                    Icon(Icons.AutoMirrored.Outlined.Assignment, contentDescription = null, tint = Brand.TextGhost, modifier = Modifier.size(28.dp))
                }
                Text("No orders yet.", color = Brand.TextMuted, fontSize = 15.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 12.dp))
                Text("Start selling!", color = Brand.TextGhost, fontSize = 13.sp)
            }
        } else {
            LazyColumn(contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)) {
                items(filtered, key = { it.idempotencyKey }) { o -> OrderRow(o, onOpenReceipt); Spacer(Modifier.size(10.dp)) }
            }
        }
    }
}

@Composable
private fun OrderRow(order: OrderRecord, onOpen: (String) -> Unit) {
    val confirmed = order.state == OrderState.CONFIRMED
    Row(
        Modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).background(Brand.Card)
            .then(if (confirmed) Modifier.clickable { onOpen(order.idempotencyKey) } else Modifier)
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        // big order number + customer + time
        Column(Modifier.padding(end = 12.dp)) {
            Text(order.orderNumberDisplay ?: "—", color = Brand.Heading, fontSize = 20.sp, fontWeight = FontWeight.ExtraBold)
            Text(order.customerName, color = Brand.TextFaint, fontSize = 11.sp, modifier = Modifier.padding(top = 3.dp))
            Text(timeOf(order.createdAtEpochMs), color = Brand.TextGhost, fontSize = 10.sp)
        }
        // middle: reference / error
        Column(Modifier.weight(1f)) {
            order.referenceId?.let { Text(it, color = Brand.TextMuted, fontSize = 11.sp, maxLines = 1) }
            order.error?.let { Text(it, color = Brand.Danger, fontSize = 11.sp) }
            if (confirmed) Text("Tap for receipt ›", color = Brand.Teal, fontSize = 11.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 2.dp))
        }
        // right: status + total
        Column(horizontalAlignment = Alignment.End) {
            val (label, c, bg) = when (order.state) {
                OrderState.PENDING -> Triple("PENDING", Brand.Packing, Brand.PackingBg)
                OrderState.CONFIRMED -> Triple("CONFIRMED", Brand.Ready, Brand.ReadyBg)
                OrderState.FAILED -> Triple("FAILED", Brand.Danger, Brand.CollectedBg)
            }
            Box(Modifier.clip(RoundedCornerShape(999.dp)).background(bg).padding(horizontal = 10.dp, vertical = 4.dp)) {
                Text(label, color = c, fontSize = 10.sp, fontWeight = FontWeight.Bold)
            }
            order.total?.let { Text("RM ${"%.2f".format(it)}", color = Brand.Heading, fontSize = 14.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(top = 6.dp)) }
        }
    }
}

private fun timeOf(epochMs: Long): String {
    val d = java.util.Date(epochMs)
    return java.text.SimpleDateFormat("HH:mm", java.util.Locale.getDefault()).format(d)
}
