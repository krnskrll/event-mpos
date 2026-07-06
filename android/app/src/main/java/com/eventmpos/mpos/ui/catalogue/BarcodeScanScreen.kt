package com.eventmpos.mpos.ui.catalogue

import com.eventmpos.mpos.core.format.rupiah

import androidx.compose.foundation.Canvas
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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.Keyboard
import androidx.compose.material.icons.outlined.ShoppingCart
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.eventmpos.mpos.core.scanner.CameraScanner
import com.eventmpos.mpos.core.scanner.PRODUCT_BARCODE_FORMATS
import com.eventmpos.mpos.core.scanner.rememberCameraPermission
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.ui.theme.Brand
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
fun BarcodeScanScreen(
    onClose: () -> Unit,
    onViewCart: () -> Unit,
    viewModel: BarcodeScanViewModel = hiltViewModel(),
) {
    val cartCount by viewModel.cartCount.collectAsStateWithLifecycle()
    val inStock by viewModel.inStock.collectAsStateWithLifecycle()
    val lookup by viewModel.lookup.collectAsStateWithLifecycle()
    val (cameraGranted, requestCamera) = rememberCameraPermission()
    val scope = rememberCoroutineScope()

    var manual by remember { mutableStateOf(false) }
    var accepting by remember { mutableStateOf(true) }
    var notFound by remember { mutableStateOf<String?>(null) }

    // Ask for the camera the moment the scanner opens.
    LaunchedEffect(Unit) { if (!cameraGranted) requestCamera() }

    // An unknown code flashes a brief message, then the scanner resumes.
    LaunchedEffect(lookup) {
        val l = lookup
        if (l is BarcodeLookup.NotFound) {
            notFound = l.code
            delay(1600)
            notFound = null
            viewModel.dismiss()
            accepting = true
        }
    }

    fun onDecoded(code: String) {
        if (!accepting) return
        accepting = false
        scope.launch { viewModel.resolve(code) }
    }

    val found = (lookup as? BarcodeLookup.Found)?.product

    Box(Modifier.fillMaxSize().background(Color(0xFF0A0A0A))) {
        // Live camera feed (or a dark "camera" backdrop when unavailable).
        if (cameraGranted) {
            CameraScanner(
                formats = PRODUCT_BARCODE_FORMATS,
                onDecoded = ::onDecoded,
                modifier = Modifier.fillMaxSize(),
            )
        } else {
            Box(
                Modifier.fillMaxSize().background(
                    Brush.radialGradient(listOf(Color(0xFF1A2A1A), Color(0xFF050A05), Color.Black)),
                ),
            )
        }

        // Top bar
        Row(
            Modifier.fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color(0xB3000000), Color.Transparent)))
                .padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Spacer(Modifier.size(24.dp))
            Text("Scan Product", color = Color.White, fontSize = 17.sp, fontWeight = FontWeight.Bold, modifier = Modifier.weight(1f), textAlign = TextAlign.Center)
            if (cartCount > 0) {
                Box(Modifier.clickable { onViewCart() }, Alignment.TopEnd) {
                    Icon(Icons.Outlined.ShoppingCart, contentDescription = "Cart", tint = Color.White, modifier = Modifier.size(22.dp).padding(end = 2.dp))
                    Box(Modifier.size(15.dp).clip(RoundedCornerShape(999.dp)).background(Brand.Teal), Alignment.Center) {
                        Text("$cartCount", color = Color.White, fontSize = 9.sp, fontWeight = FontWeight.Bold)
                    }
                }
                Spacer(Modifier.size(14.dp))
            }
            Icon(Icons.Outlined.Close, contentDescription = "Close", tint = Color.White, modifier = Modifier.size(24.dp).clickable { onClose() })
        }

        // Center scan frame with corner brackets
        Box(Modifier.fillMaxSize(), Alignment.Center) {
            if (manual) {
                ManualEntryPanel(
                    onSearch = { code -> manual = false; accepting = false; scope.launch { viewModel.resolve(code) } },
                    onBack = { manual = false; accepting = true },
                )
            } else {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    ScanFrame()
                    if (notFound != null) {
                        Spacer(Modifier.size(20.dp))
                        Text("No product for \"$notFound\"", color = Brand.Warn, fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }

        // Bottom instruction + manual / simulate affordances
        if (!manual && found == null) {
            Column(
                Modifier.align(Alignment.BottomCenter).fillMaxWidth()
                    .background(Brush.verticalGradient(listOf(Color.Transparent, Color(0xD9000000))))
                    .padding(start = 24.dp, end = 24.dp, top = 24.dp, bottom = 36.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    if (cameraGranted) "Point camera at product barcode or QR code"
                    else "Camera unavailable — enter a barcode or tap to simulate",
                    color = Color.White.copy(alpha = 0.85f), fontSize = 14.sp, textAlign = TextAlign.Center,
                )
                Spacer(Modifier.size(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(20.dp)) {
                    Row(
                        Modifier.clickable { manual = true }, verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Icon(Icons.Outlined.Keyboard, contentDescription = null, tint = Color(0xFF5BB8F5), modifier = Modifier.size(16.dp))
                        Spacer(Modifier.size(6.dp))
                        Text("Enter manually", color = Color(0xFF5BB8F5), fontSize = 14.sp)
                    }
                    if (!cameraGranted && inStock.isNotEmpty()) {
                        Text(
                            "Simulate scan",
                            color = Brand.Teal, fontSize = 14.sp, fontWeight = FontWeight.SemiBold,
                            modifier = Modifier.clickable {
                                accepting = false
                                viewModel.show(inStock[cartCount % inStock.size])
                            },
                        )
                    }
                }
            }
        }

        // Result bottom sheet
        found?.let { product ->
            ResultSheet(
                product = product,
                onDismiss = { viewModel.dismiss(); accepting = true },
                onAdd = { viewModel.addToCart(product) },
                onAdded = {
                    scope.launch {
                        delay(1200)
                        viewModel.dismiss()
                        accepting = true
                    }
                },
            )
        }
    }
}

@Composable
private fun ScanFrame() {
    Box(Modifier.size(240.dp)) {
        Canvas(Modifier.fillMaxSize()) {
            val stroke = 3.dp.toPx()
            val len = 34.dp.toPx()
            val w = size.width
            val h = size.height
            val teal = Color(0xFF17A589)
            fun l(a: Offset, b: Offset) = drawLine(teal, a, b, stroke, StrokeCap.Round)
            l(Offset(0f, 0f), Offset(len, 0f)); l(Offset(0f, 0f), Offset(0f, len))
            l(Offset(w, 0f), Offset(w - len, 0f)); l(Offset(w, 0f), Offset(w, len))
            l(Offset(0f, h), Offset(len, h)); l(Offset(0f, h), Offset(0f, h - len))
            l(Offset(w, h), Offset(w - len, h)); l(Offset(w, h), Offset(w, h - len))
        }
    }
}

@Composable
private fun ManualEntryPanel(onSearch: (String) -> Unit, onBack: () -> Unit) {
    var code by remember { mutableStateOf("") }
    Column(
        Modifier.fillMaxWidth(0.85f).clip(RoundedCornerShape(20.dp)).background(Color(0xD9000000)).padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Enter Barcode / SKU", color = Color.White, fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
        OutlinedTextField(
            value = code, onValueChange = { code = it }, singleLine = true,
            placeholder = { Text("e.g. 8990001001", color = Color.White.copy(alpha = 0.4f)) },
            keyboardOptions = KeyboardOptions(keyboardType = androidx.compose.ui.text.input.KeyboardType.Ascii),
            shape = RoundedCornerShape(12.dp),
            modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
        )
        Row(
            Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(Brand.Teal)
                .clickable { if (code.isNotBlank()) onSearch(code) }.padding(vertical = 14.dp),
            horizontalArrangement = Arrangement.Center,
        ) { Text("Search", color = Color.White, fontWeight = FontWeight.SemiBold, fontSize = 15.sp) }
        Spacer(Modifier.size(10.dp))
        Text("Back to scanner", color = Color.White.copy(alpha = 0.5f), fontSize = 13.sp, modifier = Modifier.clickable { onBack() })
    }
}

@Composable
private fun ResultSheet(product: Product, onDismiss: () -> Unit, onAdd: () -> Unit, onAdded: () -> Unit) {
    var added by remember { mutableStateOf(false) }
    Box(Modifier.fillMaxSize().background(Color(0x80000000)).clickable(enabled = !added) { onDismiss() })
    Box(Modifier.fillMaxSize(), Alignment.BottomCenter) {
        Column(
            Modifier.fillMaxWidth().clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)).background(Color.White)
                .clickable(enabled = false) {}.padding(start = 20.dp, end = 20.dp, top = 12.dp, bottom = 36.dp),
        ) {
            Box(Modifier.align(Alignment.CenterHorizontally).size(width = 40.dp, height = 4.dp).clip(RoundedCornerShape(99.dp)).background(Brand.FieldBorder))
            Spacer(Modifier.size(20.dp))
            Row {
                AsyncImage(
                    model = product.imageUrl, contentDescription = product.name, contentScale = ContentScale.Crop,
                    modifier = Modifier.size(72.dp).clip(RoundedCornerShape(12.dp)),
                )
                Spacer(Modifier.size(14.dp))
                Column(Modifier.weight(1f)) {
                    product.brand?.let { Text(it, color = Brand.TextFaint, fontSize = 12.sp) }
                    Text(product.name, color = Brand.Heading, fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.size(6.dp))
                    Text(rupiah(product.price), color = Brand.Teal, fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
            }
            Spacer(Modifier.size(16.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                val (dot, txt) = when {
                    product.outOfStock -> Brand.Danger to "Out of stock"
                    product.lowStock -> Brand.Warn to "${product.remainingStock} units available"
                    else -> Brand.Ready to "${product.remainingStock} units available"
                }
                Box(Modifier.size(8.dp).clip(RoundedCornerShape(999.dp)).background(dot))
                Spacer(Modifier.size(6.dp))
                Text(txt, color = if (product.outOfStock) Brand.Danger else Brand.TextMuted, fontSize = 13.sp)
            }
            Spacer(Modifier.size(16.dp))
            val bg = if (product.outOfStock) Brand.FieldBorder else if (added) Brand.Teal else Brand.Navy
            Row(
                Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(bg)
                    .clickable(enabled = !product.outOfStock && !added) { onAdd(); added = true; onAdded() }
                    .padding(vertical = 16.dp),
                horizontalArrangement = Arrangement.Center, verticalAlignment = Alignment.CenterVertically,
            ) {
                if (!product.outOfStock) {
                    Icon(if (added) Icons.Filled.Check else Icons.Outlined.ShoppingCart, contentDescription = null, tint = Color.White, modifier = Modifier.size(18.dp))
                    Spacer(Modifier.size(8.dp))
                }
                Text(
                    if (product.outOfStock) "Out of stock" else if (added) "Added to Cart" else "Add to Cart",
                    color = if (product.outOfStock) Brand.TextFaint else Color.White, fontWeight = FontWeight.SemiBold, fontSize = 15.sp,
                )
            }
        }
    }
}
