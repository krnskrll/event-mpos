package com.eventmpos.mpos.ui.home

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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Assignment
import androidx.compose.material.icons.outlined.ShoppingBag
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
import com.eventmpos.mpos.ui.components.Pill
import com.eventmpos.mpos.ui.components.SectionCard
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun HomeScreen(
    onCatalogue: () -> Unit,
    onOrders: () -> Unit,
    onLoggedOut: () -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val event by viewModel.event.collectAsStateWithLifecycle()
    val dash by viewModel.dashboard.collectAsStateWithLifecycle()
    val loggedOut by viewModel.loggedOut.collectAsStateWithLifecycle()
    LaunchedEffect(loggedOut) { if (loggedOut) onLoggedOut() }

    Column(modifier = Modifier.fillMaxSize().background(Brand.Bg)) {
        // Navy header
        Column(modifier = Modifier.fillMaxWidth().background(Brand.Navy).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 20.dp)) {
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
                Column(Modifier.weight(1f)) {
                    Text("Hello, ${viewModel.userName.split(" ").first()} 👋", color = Color.White, fontSize = 22.sp, fontWeight = FontWeight.Bold)
                    Text(event?.let { "${it.name} · ${it.code}" } ?: "—", color = Color.White.copy(alpha = 0.65f), fontSize = 13.sp, modifier = Modifier.padding(top = 4.dp))
                }
                Box(
                    modifier = Modifier.clip(RoundedCornerShape(20.dp)).background(Color.White.copy(alpha = 0.12f))
                        .clickable { viewModel.logout() }.padding(horizontal = 12.dp, vertical = 6.dp),
                ) { Text("End session", color = Color.White.copy(alpha = 0.85f), fontSize = 12.sp, fontWeight = FontWeight.Medium) }
            }
        }

        Column(modifier = Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(16.dp)) {
            // Catalogue CTA
            Row(
                modifier = Modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).background(Brand.Navy)
                    .clickable { onCatalogue() }.padding(20.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(Modifier.size(44.dp).clip(RoundedCornerShape(12.dp)).background(Color.White.copy(alpha = 0.15f)), Alignment.Center) {
                    Icon(Icons.Outlined.ShoppingBag, contentDescription = null, tint = Color.White, modifier = Modifier.size(22.dp))
                }
                Spacer(Modifier.size(16.dp))
                Column {
                    Text("Product Catalogue", color = Color.White, fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    Text("Browse & add items to cart", color = Color.White.copy(alpha = 0.55f), fontSize = 12.sp)
                }
            }

            Spacer(Modifier.size(12.dp))
            // My Orders Today
            SectionCard(modifier = Modifier.clickable { onOrders() }) {
                Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(36.dp).clip(RoundedCornerShape(10.dp)).background(Brand.ReadyBg), Alignment.Center) {
                        Icon(Icons.AutoMirrored.Outlined.Assignment, contentDescription = null, tint = Brand.Ready, modifier = Modifier.size(18.dp))
                    }
                    Spacer(Modifier.size(10.dp))
                    Column(Modifier.weight(1f)) {
                        Text("My Orders Today", color = Brand.Text, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                        Text("Session summary", color = Brand.TextFaint, fontSize = 12.sp)
                    }
                    Pill("${dash.orderCount}", Brand.Navy, Brand.NavActiveBg)
                }
                Spacer(Modifier.size(12.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    if (dash.confirmed > 0) Pill("${dash.confirmed} Confirmed", Brand.Ready, Brand.ReadyBg)
                    if (dash.pending > 0) Pill("${dash.pending} Pending", Brand.Packing, Brand.PackingBg)
                    if (dash.failed > 0) Pill("${dash.failed} Failed", Brand.Danger, Brand.CollectedBg)
                    if (dash.orderCount == 0) Text("No orders yet this session.", color = Brand.TextGhost, fontSize = 12.sp)
                }
            }

            Spacer(Modifier.size(12.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                StatCard(Modifier.weight(1f), "Revenue Today", "RM ${"%.2f".format(dash.revenue)}", Brand.Teal)
                StatCard(Modifier.weight(1f), "Orders", "${dash.orderCount}", Brand.Navy)
            }
        }
    }
}

@Composable
private fun StatCard(modifier: Modifier, label: String, value: String, valueColor: Color) {
    SectionCard(modifier = modifier, padding = androidx.compose.foundation.layout.PaddingValues(14.dp)) {
        Text(label, color = Brand.TextFaint, fontSize = 11.sp)
        Text(value, color = valueColor, fontSize = 16.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(top = 4.dp))
    }
}
