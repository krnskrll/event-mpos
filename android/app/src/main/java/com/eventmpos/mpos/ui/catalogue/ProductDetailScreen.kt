package com.eventmpos.mpos.ui.catalogue

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.outlined.ShoppingCart
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.ui.graphics.vector.ImageVector
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.eventmpos.mpos.ui.components.BrandButton
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

    val p = product
    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        Box {
            AsyncImage(
                model = p?.imageUrl, contentDescription = p?.name, contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxWidth().height(280.dp),
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
        Column(modifier = Modifier.fillMaxSize().padding(20.dp)) {
            p.brand?.let { Text(it, color = Brand.TextFaint, fontSize = 13.sp) }
            Row(Modifier.fillMaxWidth().padding(top = 4.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.Top) {
                Text(p.name, color = Brand.Heading, fontSize = 20.sp, fontWeight = FontWeight.Bold, modifier = Modifier.weight(1f))
                Text("RM ${"%.2f".format(p.price)}", color = Brand.Teal, fontSize = 20.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(start = 8.dp))
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

            Spacer(Modifier.weight(1f))

            // qty stepper
            Row(Modifier.fillMaxWidth().padding(bottom = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                Text("Quantity", color = Color(0xFF374151), fontSize = 14.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.weight(1f))
                StepBtn("−") { if (qty > 1) qty-- }
                Text("$qty", modifier = Modifier.padding(horizontal = 16.dp), fontWeight = FontWeight.Bold, fontSize = 16.sp)
                StepBtn("+") { qty++ }
            }
            BrandButton(
                text = if (p.outOfStock) "Out of stock" else if (added) "Added ✓" else "Add to cart · RM ${"%.2f".format(p.price * qty)}",
                onClick = { viewModel.addToCart(qty); added = true },
                enabled = !p.outOfStock,
                color = if (added) Brand.Ready else Brand.Teal,
                modifier = Modifier.fillMaxWidth(),
            )
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
private fun StepBtn(label: String, onClick: () -> Unit) {
    Box(Modifier.size(40.dp).clip(RoundedCornerShape(10.dp)).background(Brand.CardBorder).clickable { onClick() }, Alignment.Center) {
        Text(label, fontSize = 20.sp, color = Brand.Navy, fontWeight = FontWeight.Bold)
    }
}
