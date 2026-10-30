package com.eventmpos.mpos.ui.cart

import com.eventmpos.mpos.core.format.rupiah

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
import androidx.compose.material.icons.outlined.Check
import androidx.compose.material.icons.outlined.PersonOutline
import androidx.compose.material3.HorizontalDivider
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.components.ProductImage
import com.eventmpos.mpos.ui.components.SectionCard
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun CartScreen(
    onConfirmed: () -> Unit,
    onIdentifyCustomer: () -> Unit,
    viewModel: CartViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val drafts by viewModel.drafts.collectAsStateWithLifecycle()
    var showSave by remember { mutableStateOf(false) }
    var showDrafts by remember { mutableStateOf(false) }

    if (showSave) SaveDraftDialog(state.customer?.name.orEmpty(), { viewModel.saveDraft(it); showSave = false }, { showSave = false })
    if (showDrafts) DraftsDialog(drafts, state.lines.isEmpty(), viewModel::resumeDraft, viewModel::deleteDraft) { showDrafts = false }

    Column(modifier = Modifier.fillMaxSize().background(Brand.Bg)) {
        Column(Modifier.fillMaxWidth().background(Brand.Navy).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f)) {
                    Text("Cart", color = Color.White, fontSize = 20.sp, fontWeight = FontWeight.Bold)
                    Text("${state.lines.sumOf { it.qty }} items", color = Color.White.copy(alpha = 0.6f), fontSize = 13.sp)
                }
                if (state.lines.isNotEmpty()) HeaderAction("Save draft") { showSave = true }
                HeaderAction("Drafts (${drafts.size})") { showDrafts = true }
            }
        }

        if (state.lines.isEmpty()) {
            Box(Modifier.fillMaxSize(), Alignment.Center) {
                Text("Cart is empty.\nAdd products from the catalogue.", color = Brand.TextMuted, textAlign = androidx.compose.ui.text.style.TextAlign.Center)
            }
        } else {
        LazyColumn(Modifier.weight(1f), contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)) {
            items(state.lines, key = { it.key }) { line ->
                SectionCard(modifier = Modifier.padding(bottom = 10.dp), padding = androidx.compose.foundation.layout.PaddingValues(10.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        ProductImage(
                            imageUrl = line.product.imageUrl, contentDescription = line.product.name,
                            modifier = Modifier.size(52.dp).clip(RoundedCornerShape(10.dp)),
                        )
                        Spacer(Modifier.size(10.dp))
                        Column(Modifier.weight(1f)) {
                            Text(line.displayName, color = Brand.Text, fontSize = 13.sp, fontWeight = FontWeight.SemiBold, maxLines = 2)
                            Text(rupiah(line.lineTotal), color = Brand.Teal, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                        }
                        StepBtn("−") { viewModel.decrement(line.key) }
                        Text("${line.qty}", modifier = Modifier.padding(horizontal = 12.dp), fontWeight = FontWeight.Bold)
                        StepBtn("+") { viewModel.increment(line.key) }
                    }
                }
            }
        }

        // footer
        SectionCard(modifier = Modifier.padding(16.dp)) {
            Row(Modifier.fillMaxWidth().clickable { onIdentifyCustomer() }, verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(32.dp).clip(CircleShape).background(if (state.customer != null) Brand.TealSoft else Brand.CardBorder), Alignment.Center) {
                    Icon(
                        if (state.customer != null) Icons.Outlined.Check else Icons.Outlined.PersonOutline,
                        contentDescription = null, tint = if (state.customer != null) Brand.Teal else Brand.TextMuted, modifier = Modifier.size(18.dp),
                    )
                }
                Spacer(Modifier.size(10.dp))
                Text(
                    state.customer?.let { "${it.name}${it.tier?.let { t -> " · $t" } ?: ""}" } ?: "Identify customer",
                    color = if (state.customer != null) Brand.Text else Brand.Danger, fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.weight(1f),
                )
                Text(if (state.customer != null) "Change" else "Add", color = Brand.Teal, fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
            }
            HorizontalDivider(Modifier.padding(vertical = 12.dp), color = Brand.CardBorder)
            TotalRow("Subtotal", state.subtotal, Brand.TextMuted)
            if (state.discount > 0.0) TotalRow("Discount (${state.promotionName ?: "promo"})", -state.discount, Brand.Teal)
            Spacer(Modifier.size(4.dp))
            TotalRow("Total", state.total, Brand.Heading, bold = true)
            Spacer(Modifier.size(12.dp))
            BrandButton(
                text = when {
                    state.submitting -> "Submitting…"
                    state.customer == null -> "Identify a customer to confirm"
                    else -> "Confirm order"
                },
                onClick = { viewModel.confirm(onConfirmed) },
                enabled = viewModel.canConfirm && !state.submitting,
                loading = state.submitting,
                modifier = Modifier.fillMaxWidth(),
            )
        }
        }
    }
}

@Composable
private fun HeaderAction(label: String, onClick: () -> Unit) {
    Text(
        label, color = Color.White, fontSize = 13.sp, fontWeight = FontWeight.SemiBold,
        modifier = Modifier.clip(RoundedCornerShape(8.dp)).clickable { onClick() }.padding(horizontal = 10.dp, vertical = 6.dp),
    )
}

@Composable
private fun StepBtn(label: String, onClick: () -> Unit) {
    Box(Modifier.size(34.dp).clip(RoundedCornerShape(9.dp)).background(Brand.CardBorder).clickable { onClick() }, Alignment.Center) {
        Text(label, fontSize = 18.sp, color = Brand.Navy, fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun TotalRow(label: String, amount: Double, color: Color, bold: Boolean = false) {
    Row(Modifier.fillMaxWidth().padding(vertical = 2.dp), Arrangement.SpaceBetween) {
        Text(label, color = color, fontWeight = if (bold) FontWeight.Bold else FontWeight.Normal, fontSize = if (bold) 16.sp else 14.sp)
        Text(rupiah(amount), color = color, fontWeight = if (bold) FontWeight.Bold else FontWeight.Normal, fontSize = if (bold) 16.sp else 14.sp)
    }
}
