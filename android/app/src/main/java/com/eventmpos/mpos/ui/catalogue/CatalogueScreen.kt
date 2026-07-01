package com.eventmpos.mpos.ui.catalogue

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.BasicAlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
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
import coil.compose.AsyncImage
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.components.Pill
import com.eventmpos.mpos.ui.theme.Brand

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CatalogueScreen(
    onProductSelect: (String) -> Unit,
    onCustomerScan: () -> Unit,
    viewModel: CatalogueViewModel = hiltViewModel(),
) {
    val products by viewModel.products.collectAsStateWithLifecycle()
    val query by viewModel.searchQuery.collectAsStateWithLifecycle()
    val categories by viewModel.categories.collectAsStateWithLifecycle()
    val category by viewModel.category.collectAsStateWithLifecycle()
    val customer by viewModel.customer.collectAsStateWithLifecycle()
    val scan by viewModel.scan.collectAsStateWithLifecycle()
    val snackbar = remember { SnackbarHostState() }
    var showScan by remember { mutableStateOf(false) }

    LaunchedEffect(scan) {
        when (val s = scan) {
            is ScanFeedback.Found -> snackbar.showSnackbar("Added: ${s.product.name}")
            is ScanFeedback.NotFound -> snackbar.showSnackbar("No product for ${s.barcode}")
            null -> {}
        }
        viewModel.clearScan()
    }

    Column(modifier = Modifier.fillMaxSize().background(Color.White)) {
        // top bar
        Column(modifier = Modifier.fillMaxWidth().padding(top = 40.dp, start = 16.dp, end = 16.dp, bottom = 12.dp)) {
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f)) {
                    Text("Current Event", color = Brand.TextFaint, fontSize = 11.sp)
                    Text(viewModel.eventName, color = Brand.Heading, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                }
                Text("● Synced", color = Brand.Teal, fontSize = 12.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(Modifier.size(12.dp))
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = query,
                    onValueChange = viewModel::onSearch,
                    placeholder = { Text("Search product…", color = Brand.TextFaint) },
                    singleLine = true,
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.weight(1f),
                )
                IconBox(bg = Brand.Navy, label = "⌗", onClick = { showScan = true })
                IconBox(
                    bg = if (customer != null) Brand.TealSoft else Brand.CardBorder,
                    label = "▣",
                    tint = if (customer != null) Brand.Teal else Brand.TextMuted,
                    onClick = onCustomerScan,
                )
            }
            if (customer != null) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(top = 10.dp).clip(RoundedCornerShape(10.dp))
                        .background(Brand.TealSoft).border(1.dp, Brand.TealBorder, RoundedCornerShape(10.dp)).padding(10.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("✓ Customer linked: ", color = Brand.Ready, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                    Text(customer!!.name, color = Color(0xFF374151), fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
                }
            }
        }

        // category pills
        LazyRow(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 12.dp, vertical = 8.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            items(categories) { cat ->
                val selected = cat == category
                Box(
                    modifier = Modifier.clip(RoundedCornerShape(999.dp)).background(if (selected) Brand.Navy else Brand.CardBorder)
                        .clickable { viewModel.onCategory(cat) }.padding(horizontal = 16.dp, vertical = 6.dp),
                ) { Text(cat, color = if (selected) Color.White else Brand.TextMuted, fontSize = 13.sp, fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal) }
            }
        }

        Box(Modifier.fillMaxSize()) {
            LazyVerticalGrid(
                columns = GridCells.Fixed(2),
                contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
                modifier = Modifier.fillMaxSize(),
            ) {
                items(products, key = { it.id }) { p -> ProductCard(p) { onProductSelect(p.id) } }
            }
            SnackbarHost(snackbar, modifier = Modifier.align(Alignment.BottomCenter))
        }
    }

    if (showScan) {
        BasicAlertDialog(onDismissRequest = { showScan = false }) {
            Column(Modifier.clip(RoundedCornerShape(16.dp)).background(Color.White).padding(20.dp)) {
                Text("Scan / enter barcode", color = Brand.Heading, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                var code by remember { mutableStateOf("") }
                OutlinedTextField(
                    value = code, onValueChange = { code = it }, singleLine = true,
                    placeholder = { Text("e.g. 8990001001") },
                    keyboardOptions = KeyboardOptions(keyboardType = androidx.compose.ui.text.input.KeyboardType.Number),
                    modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
                )
                BrandButton("Add to cart", onClick = { viewModel.submitManualBarcode(code); showScan = false }, color = Brand.Teal, modifier = Modifier.fillMaxWidth())
            }
        }
    }
}

@Composable
private fun IconBox(bg: Color, label: String, tint: Color = Color.White, onClick: () -> Unit) {
    Box(Modifier.size(48.dp).clip(RoundedCornerShape(12.dp)).background(bg).clickable { onClick() }, Alignment.Center) {
        Text(label, color = tint, fontSize = 18.sp)
    }
}

@Composable
private fun ProductCard(product: Product, onClick: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxWidth().clip(RoundedCornerShape(16.dp)).background(Brand.Card)
            .clickable(enabled = !product.outOfStock) { onClick() }
            .let { if (product.outOfStock) it.then(Modifier.background(Color.White.copy(alpha = 0.4f))) else it },
    ) {
        Box {
            AsyncImage(
                model = product.imageUrl, contentDescription = product.name, contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxWidth().aspectRatio(1f).clip(RoundedCornerShape(topStart = 16.dp, topEnd = 16.dp)),
            )
            val (txt, c, bg) = when {
                product.outOfStock -> Triple("Out of stock", Color.White, Brand.Danger)
                product.lowStock -> Triple("${product.remainingStock} left", Color.White, Brand.Warn)
                else -> Triple("${product.remainingStock} left", Brand.TextMuted, Color.White)
            }
            Pill(txt, c, bg, modifier = Modifier.align(Alignment.TopEnd).padding(8.dp))
        }
        Column(Modifier.padding(10.dp)) {
            product.brand?.let { Text(it, color = Brand.TextFaint, fontSize = 11.sp) }
            Text(product.name, color = Brand.Text, fontSize = 13.sp, fontWeight = FontWeight.SemiBold, maxLines = 2, modifier = Modifier.padding(vertical = 2.dp))
            Text("RM ${"%.2f".format(product.price)}", color = Brand.Teal, fontSize = 13.sp, fontWeight = FontWeight.Bold)
        }
    }
}
