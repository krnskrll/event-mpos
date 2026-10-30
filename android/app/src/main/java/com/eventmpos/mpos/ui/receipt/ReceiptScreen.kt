package com.eventmpos.mpos.ui.receipt

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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun ReceiptScreen(
    onNewOrder: () -> Unit,
    onViewOrders: () -> Unit,
    viewModel: ReceiptViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().background(Brand.Bg).padding(20.dp), horizontalAlignment = Alignment.CenterHorizontally) {
        Spacer(Modifier.size(12.dp))
        Box(Modifier.size(64.dp).clip(CircleShape).background(Brand.TealSoft), Alignment.Center) {
            Text("✓", color = Brand.Teal, fontSize = 34.sp)
        }
        Text(if (state.cancelled) "Order cancelled" else "Order confirmed", color = Brand.Heading, fontSize = 20.sp, fontWeight = androidx.compose.ui.text.font.FontWeight.Bold, modifier = Modifier.padding(top = 12.dp))
        Text("Receipt printed automatically", color = Brand.TextMuted, fontSize = 13.sp)

        when {
            state.loading -> { Spacer(Modifier.size(24.dp)); CircularProgressIndicator(color = Brand.Navy) }
            state.lines.isEmpty() -> Text(state.message ?: "No receipt.", color = Brand.Danger, modifier = Modifier.padding(top = 24.dp))
            else -> {
                Column(
                    modifier = Modifier.fillMaxWidth().weight(1f).padding(top = 16.dp).clip(RoundedCornerShape(12.dp)).background(Color.White)
                        .verticalScroll(rememberScrollState()).padding(16.dp),
                ) {
                    Text(state.lines.joinToString("\n"), fontFamily = FontFamily.Monospace, fontSize = 12.sp, color = Brand.Text)
                }
                state.message?.let {
                    Text(it, color = if (state.error) Brand.Danger else Brand.Teal, fontSize = 13.sp, modifier = Modifier.padding(top = 8.dp))
                }
                Spacer(Modifier.size(12.dp))
                BrandButton(if (state.printing) "Printing…" else "Reprint receipt", onClick = viewModel::reprint, enabled = !state.printing, color = Brand.Navy, modifier = Modifier.fillMaxWidth())
                Row(Modifier.fillMaxWidth().padding(top = 8.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    BrandButton("New order", onClick = onNewOrder, color = Brand.Teal, modifier = Modifier.weight(1f))
                    BrandButton("View orders", onClick = onViewOrders, color = Brand.TextMuted, modifier = Modifier.weight(1f))
                }
            }
        }
    }
}
