package com.eventmpos.mpos.ui.catalogue

import com.eventmpos.mpos.core.format.rupiah

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.material.icons.outlined.ShoppingCart
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.ui.components.ProductImage
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun ProductDetailScreen(
    onBack: () -> Unit,
    onViewCart: () -> Unit,
    viewModel: ProductDetailViewModel = hiltViewModel(),
) {
    val product by viewModel.product.collectAsStateWithLifecycle()
    val cartCount by viewModel.cartCount.collectAsStateWithLifecycle()
    var qty by remember { mutableIntStateOf(1) }
    var added by remember { mutableStateOf(false) }
    var descOpen by remember { mutableStateOf(false) }

    val p = product
    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        // Hero image with overlay nav
        Box {
            ProductImage(
                imageUrl = p?.imageUrl, contentDescription = p?.name,
                modifier = Modifier.fillMaxWidth().height(260.dp),
            )
            Row(Modifier.fillMaxWidth().padding(top = 40.dp, start = 16.dp, end = 16.dp), horizontalArrangement = Arrangement.SpaceBetween) {
                CircleBtn(Icons.AutoMirrored.Filled.ArrowBack, onBack)
                Box {
                    CircleBtn(Icons.Outlined.ShoppingCart, onViewCart)
                    if (cartCount > 0) {
                        Box(Modifier.align(Alignment.TopEnd).size(16.dp).clip(CircleShape).background(Brand.Teal), Alignment.Center) {
                            Text("$cartCount", color = Color.White, fontSize = 9.sp, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }

        if (p == null) {
            Box(Modifier.fillMaxSize(), Alignment.Center) { Text("Loading…", color = Brand.TextMuted) }
        } else {
            // Detail sheet (overlaps hero, rounded top) — scrollable
            Column(
                modifier = Modifier.weight(1f).fillMaxWidth().offset(y = (-20).dp)
                    .clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)).background(Color.White)
                    .verticalScroll(rememberScrollState()).padding(20.dp),
            ) {
                p.brand?.let { Text(it, color = Brand.TextFaint, fontSize = 13.sp) }
                Row(Modifier.fillMaxWidth().padding(top = 4.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.Top) {
                    Text(p.name, color = Brand.Heading, fontSize = 20.sp, fontWeight = FontWeight.Bold, modifier = Modifier.weight(1f))
                    Text(rupiah(p.price), color = Brand.Teal, fontSize = 20.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(start = 8.dp))
                }
                Spacer(Modifier.height(10.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    val (dot, txt) = when {
                        p.outOfStock -> Brand.Danger to "Out of stock"
                        p.lowStock -> Brand.Warn to "Low stock: ${p.remainingStock} units"
                        else -> Brand.Ready to "In stock: ${p.remainingStock} units"
                    }
                    Box(Modifier.size(8.dp).clip(CircleShape).background(dot))
                    Spacer(Modifier.size(6.dp))
                    Text(txt, color = dot, fontSize = 13.sp, fontWeight = FontWeight.Medium)
                }
                Text("SKU: ${p.sku}", color = Brand.TextGhost, fontSize = 11.sp, modifier = Modifier.padding(top = 8.dp))

                // Quantity selector
                Text("Quantity", color = Color(0xFF374151), fontSize = 13.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 20.dp, bottom = 8.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    QtyBtn("−", enabled = qty > 1) { if (qty > 1) qty-- }
                    Text("$qty", modifier = Modifier.padding(horizontal = 16.dp).width(28.dp), fontWeight = FontWeight.Bold, fontSize = 20.sp, color = Brand.Navy, textAlign = androidx.compose.ui.text.style.TextAlign.Center)
                    QtyBtn("+", enabled = !p.outOfStock && qty < p.remainingStock) { if (qty < p.remainingStock) qty++ }
                    Spacer(Modifier.size(12.dp))
                    Text(if (p.outOfStock) "Unavailable" else "max ${p.remainingStock}", color = Brand.TextFaint, fontSize = 13.sp)
                }

                // Description accordion
                Column(Modifier.padding(top = 20.dp).clip(RoundedCornerShape(12.dp)).border(1.dp, Brand.CardBorder, RoundedCornerShape(12.dp))) {
                    Row(
                        Modifier.fillMaxWidth().clickable { descOpen = !descOpen }.padding(horizontal = 16.dp, vertical = 12.dp),
                        horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("Details", color = Color(0xFF374151), fontSize = 14.sp, fontWeight = FontWeight.SemiBold)
                        Icon(if (descOpen) Icons.Filled.KeyboardArrowUp else Icons.Filled.KeyboardArrowDown, contentDescription = null, tint = Brand.TextFaint)
                    }
                    AnimatedVisibility(descOpen) {
                        Column(Modifier.padding(start = 16.dp, end = 16.dp, bottom = 14.dp)) {
                            Text("Brand: ${p.brand ?: "—"}", color = Brand.TextMuted, fontSize = 13.sp)
                            Text("SKU: ${p.sku}", color = Brand.TextMuted, fontSize = 13.sp, modifier = Modifier.padding(top = 4.dp))
                            if (p.barcodes.isNotEmpty()) {
                                Text("Barcodes: ${p.barcodes.joinToString(", ")}", color = Brand.TextMuted, fontSize = 13.sp, modifier = Modifier.padding(top = 4.dp))
                            }
                        }
                    }
                }
                Spacer(Modifier.height(8.dp))
            }

            // Fixed footer CTA
            Box(
                Modifier.fillMaxWidth().background(Color.White)
                    .border(width = 1.dp, color = Brand.CardBorder, shape = RoundedCornerShape(0.dp))
                    .padding(20.dp),
            ) {
                val bg = if (p.outOfStock) Brand.FieldBorder else if (added) Brand.Teal else Brand.Navy
                Row(
                    Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(bg)
                        .clickable(enabled = !p.outOfStock) { viewModel.addToCart(qty); added = true }
                        .padding(vertical = 16.dp),
                    horizontalArrangement = Arrangement.Center, verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (!p.outOfStock) {
                        Icon(if (added) Icons.Filled.Check else Icons.Outlined.ShoppingCart, contentDescription = null, tint = Color.White, modifier = Modifier.size(18.dp))
                        Spacer(Modifier.size(8.dp))
                    }
                    Text(
                        if (p.outOfStock) "Out of Stock" else if (added) "Added to Cart" else "Add to Cart · ${rupiah(p.price * qty)}",
                        color = if (p.outOfStock) Brand.TextFaint else Color.White, fontWeight = FontWeight.SemiBold, fontSize = 16.sp,
                    )
                }
            }
        }
    }
}

@Composable
private fun CircleBtn(icon: ImageVector, onClick: () -> Unit) {
    Box(Modifier.size(36.dp).clip(CircleShape).background(Color.White.copy(alpha = 0.9f)).clickable { onClick() }, Alignment.Center) {
        Icon(icon, contentDescription = null, tint = Brand.Navy, modifier = Modifier.size(18.dp))
    }
}

@Composable
private fun QtyBtn(label: String, enabled: Boolean, onClick: () -> Unit) {
    Box(
        Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(if (enabled) Brand.NavActiveBg else Brand.CardBorder).clickable(enabled = enabled) { onClick() },
        Alignment.Center,
    ) { Text(label, fontSize = 20.sp, color = if (enabled) Brand.Navy else Brand.TextGhost, fontWeight = FontWeight.Bold) }
}
