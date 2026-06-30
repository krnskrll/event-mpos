package com.eventmpos.mpos.ui.catalogue

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.eventmpos.mpos.domain.product.Product

/**
 * Browsable product grid for the selected event. Data comes from Room (works offline). A scan (or
 * typed barcode) resolves to a product via the ScannerPort. Adding to a cart arrives in EPIC-07.
 */
@Composable
fun CatalogueScreen(
    onViewCart: (String) -> Unit,
    viewModel: CatalogueViewModel = hiltViewModel(),
) {
    val products by viewModel.products.collectAsStateWithLifecycle()
    val query by viewModel.searchQuery.collectAsStateWithLifecycle()
    val scan by viewModel.scan.collectAsStateWithLifecycle()
    val cartCount by viewModel.cartCount.collectAsStateWithLifecycle()
    val snackbar = remember { SnackbarHostState() }

    LaunchedEffect(scan) {
        when (val s = scan) {
            is ScanFeedback.Found -> snackbar.showSnackbar("Scanned: ${s.product.name} (RM ${"%.2f".format(s.product.price)})")
            is ScanFeedback.NotFound -> snackbar.showSnackbar("No product for barcode ${s.barcode}")
            null -> {}
        }
        viewModel.clearScan()
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbar) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(12.dp)) {
            androidx.compose.foundation.layout.Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = androidx.compose.foundation.layout.Arrangement.SpaceBetween,
            ) {
                Text("Product catalogue", style = MaterialTheme.typography.titleLarge)
                androidx.compose.material3.Button(onClick = { onViewCart(viewModel.eventId) }) {
                    Text("View cart ($cartCount)")
                }
            }

            OutlinedTextField(
                value = query,
                onValueChange = viewModel::onSearch,
                label = { Text("Search or type a barcode") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
            )
            // Pressing the action also resolves the typed value as a "scan" (emulator-friendly).
            androidx.compose.material3.TextButton(onClick = { viewModel.submitManualBarcode(query) }) {
                Text("Scan typed barcode")
            }

            if (products.isEmpty()) {
                Text("No products synced for this event yet.", color = MaterialTheme.colorScheme.onSurfaceVariant)
            } else {
                LazyVerticalGrid(
                    columns = GridCells.Fixed(2),
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.fillMaxSize(),
                ) {
                    items(products, key = { it.id }) { product ->
                        ProductCard(product, onAdd = { viewModel.addToCart(product) })
                    }
                }
            }
        }
    }
}

@Composable
private fun StockLabel(product: Product) {
    val (text, color) = when {
        product.outOfStock -> "Sold out" to MaterialTheme.colorScheme.error
        product.lowStock -> "Low stock · ${product.remainingStock} left" to MaterialTheme.colorScheme.error
        else -> "${product.remainingStock} in stock" to MaterialTheme.colorScheme.onSurfaceVariant
    }
    Text(text, style = MaterialTheme.typography.bodySmall, color = color, modifier = Modifier.padding(top = 2.dp))
}

@Composable
private fun ProductCard(product: Product, onAdd: () -> Unit) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column {
            AsyncImage(
                model = product.imageUrl,
                contentDescription = product.name,
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxWidth().aspectRatio(1f),
            )
            Column(Modifier.padding(10.dp)) {
                Text(product.name, fontWeight = FontWeight.SemiBold, style = MaterialTheme.typography.bodyMedium)
                product.brand?.let {
                    Text(it, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                Text("RM ${"%.2f".format(product.price)}", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 4.dp))
                StockLabel(product)
                androidx.compose.material3.Button(
                    onClick = onAdd,
                    enabled = !product.outOfStock,
                    modifier = Modifier.fillMaxWidth().padding(top = 6.dp),
                ) {
                    Text(if (product.outOfStock) "Sold out" else "Add")
                }
            }
        }
    }
}
