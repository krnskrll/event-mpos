package com.eventmpos.mpos.ui.customer

import androidx.compose.foundation.Canvas
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
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
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.core.scanner.CameraScanner
import com.eventmpos.mpos.core.scanner.MEMBER_QR_FORMATS
import com.eventmpos.mpos.core.scanner.rememberCameraPermission
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.SelectedCustomer
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun CustomerScreen(
    onBack: () -> Unit,
    onDone: () -> Unit,
    viewModel: CustomerViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val (cameraGranted, requestCamera) = rememberCameraPermission()

    var manual by remember { mutableStateOf(false) }
    var accepting by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) { if (!cameraGranted) requestCamera() }

    // A "not found" QR drops the BA into the manual panel automatically.
    LaunchedEffect(state.notFoundQr) { if (state.notFoundQr != null) { manual = true; accepting = true } }

    fun onDecoded(code: String) {
        if (!accepting || state.identifying) return
        accepting = false
        viewModel.identifyByQr(code)
    }

    val selected = state.selected

    Box(Modifier.fillMaxSize().background(Color(0xFF06060F))) {
        if (cameraGranted && selected == null && !manual) {
            CameraScanner(formats = MEMBER_QR_FORMATS, onDecoded = ::onDecoded, modifier = Modifier.fillMaxSize())
        } else if (selected == null && !manual) {
            Box(Modifier.fillMaxSize().background(Brush.radialGradient(listOf(Color(0xFF16213A), Color(0xFF06060F), Color.Black))))
        }

        // Top bar
        Row(
            Modifier.fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color(0xB3000000), Color.Transparent)))
                .padding(top = 44.dp, start = 16.dp, end = 20.dp, bottom = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back", tint = Color.White, modifier = Modifier.size(24.dp).clickable { onBack() })
            Text("Scan Member QR", color = Color.White, fontSize = 17.sp, fontWeight = FontWeight.Bold, modifier = Modifier.weight(1f).padding(start = 12.dp))
            Text("Skip", color = Color.White.copy(alpha = 0.7f), fontSize = 14.sp, modifier = Modifier.clickable { onBack() })
        }

        // Center content
        Box(Modifier.fillMaxSize(), Alignment.Center) {
            when {
                state.identifying -> CircularProgressIndicator(color = Brand.Teal)
                manual && selected == null -> ManualPanel(
                    notFoundQr = state.notFoundQr,
                    onIdentify = { code -> accepting = false; viewModel.identifyByQr(code) },
                    onUseName = { name -> viewModel.setManual(name) },
                    onBackToScanner = { manual = false; accepting = true },
                )
                selected == null -> QrFrame()
            }
        }

        // Bottom instruction
        if (selected == null && !manual && !state.identifying) {
            Column(
                Modifier.align(Alignment.BottomCenter).fillMaxWidth()
                    .background(Brush.verticalGradient(listOf(Color.Transparent, Color(0xD9000000))))
                    .padding(start = 24.dp, end = 24.dp, top = 24.dp, bottom = 36.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    if (cameraGranted) "Point camera at the member's QR code"
                    else "Camera unavailable — enter the member ID",
                    color = Color.White.copy(alpha = 0.85f), fontSize = 14.sp, textAlign = TextAlign.Center,
                )
                Spacer(Modifier.size(12.dp))
                Text("Enter member ID / name manually", color = Color(0xFF5BB8F5), fontSize = 14.sp, modifier = Modifier.clickable { manual = true })
            }
        }

        // Profile result sheet
        selected?.let { customer ->
            ProfileSheet(
                customer = customer,
                onConfirm = onDone,
                onScanAgain = { viewModel.clear(); accepting = true; manual = false },
            )
        }
    }
}

@Composable
private fun QrFrame() {
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
private fun ManualPanel(
    notFoundQr: String?,
    onIdentify: (String) -> Unit,
    onUseName: (String) -> Unit,
    onBackToScanner: () -> Unit,
) {
    var code by remember { mutableStateOf("") }
    var name by remember { mutableStateOf("") }
    Column(
        Modifier.fillMaxWidth(0.88f).clip(RoundedCornerShape(20.dp)).background(Color(0xD9000000)).padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Identify Member", color = Color.White, fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
        if (notFoundQr != null) {
            Spacer(Modifier.size(6.dp))
            Text("No member for \"$notFoundQr\"", color = Brand.Warn, fontSize = 12.sp)
        }
        OutlinedTextField(
            value = code, onValueChange = { code = it }, singleLine = true,
            placeholder = { Text("Member ID / QR (e.g. MBR-00001)", color = Color.White.copy(alpha = 0.4f)) },
            shape = RoundedCornerShape(12.dp),
            modifier = Modifier.fillMaxWidth().padding(top = 12.dp),
        )
        Row(
            Modifier.fillMaxWidth().padding(top = 10.dp).clip(RoundedCornerShape(12.dp)).background(Brand.Teal)
                .clickable { if (code.isNotBlank()) onIdentify(code) }.padding(vertical = 14.dp),
            horizontalArrangement = Arrangement.Center,
        ) { Text("Look up", color = Color.White, fontWeight = FontWeight.SemiBold, fontSize = 15.sp) }

        Spacer(Modifier.size(16.dp))
        Text("or continue with just a name", color = Color.White.copy(alpha = 0.5f), fontSize = 12.sp)
        OutlinedTextField(
            value = name, onValueChange = { name = it }, singleLine = true,
            placeholder = { Text("Customer name", color = Color.White.copy(alpha = 0.4f)) },
            shape = RoundedCornerShape(12.dp),
            modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
        )
        Row(
            Modifier.fillMaxWidth().padding(top = 10.dp).clip(RoundedCornerShape(12.dp)).background(Brand.Navy)
                .clickable { if (name.isNotBlank()) onUseName(name) }.padding(vertical = 14.dp),
            horizontalArrangement = Arrangement.Center,
        ) { Text("Use this name", color = Color.White, fontWeight = FontWeight.SemiBold, fontSize = 15.sp) }

        Spacer(Modifier.size(10.dp))
        Text("Back to scanner", color = Color.White.copy(alpha = 0.5f), fontSize = 13.sp, modifier = Modifier.clickable { onBackToScanner() })
    }
}

@Composable
private fun ProfileSheet(customer: SelectedCustomer, onConfirm: () -> Unit, onScanAgain: () -> Unit) {
    Box(Modifier.fillMaxSize().background(Color(0x80000000)))
    Box(Modifier.fillMaxSize(), Alignment.BottomCenter) {
        Column(
            Modifier.fillMaxWidth().clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp)).background(Color.White)
                .padding(start = 24.dp, end = 24.dp, top = 12.dp, bottom = 36.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Box(Modifier.size(width = 40.dp, height = 4.dp).clip(RoundedCornerShape(99.dp)).background(Brand.FieldBorder))
            Spacer(Modifier.size(20.dp))
            Box(Modifier.size(64.dp).clip(CircleShape).background(Brand.TealSoft), Alignment.Center) {
                Icon(Icons.Filled.CheckCircle, contentDescription = null, tint = Brand.Teal, modifier = Modifier.size(36.dp))
            }
            Spacer(Modifier.size(14.dp))
            Text(customer.name, color = Brand.Heading, fontSize = 20.sp, fontWeight = FontWeight.Bold)
            Spacer(Modifier.size(6.dp))
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                customer.tier?.let {
                    Box(Modifier.clip(RoundedCornerShape(999.dp)).background(Brand.GoldBg).padding(horizontal = 10.dp, vertical = 3.dp)) {
                        Text(it, color = Brand.Gold, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    }
                }
                Text(
                    when (customer.source) {
                        CustomerSource.LOCAL_CACHE -> "matched offline"
                        CustomerSource.ONLINE_LOOKUP -> "matched online"
                        CustomerSource.MANUAL -> "manually entered"
                    },
                    color = Brand.TextFaint, fontSize = 12.sp,
                )
            }
            customer.qrCode?.let {
                Spacer(Modifier.size(4.dp))
                Text(it, color = Brand.TextMuted, fontSize = 12.sp)
            }
            Spacer(Modifier.size(20.dp))
            Row(
                Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(Brand.Navy)
                    .clickable { onConfirm() }.padding(vertical = 16.dp),
                horizontalArrangement = Arrangement.Center,
            ) { Text("Confirm & Continue", color = Color.White, fontWeight = FontWeight.SemiBold, fontSize = 15.sp) }
            Spacer(Modifier.size(10.dp))
            Text("Scan a different member", color = Brand.Teal, fontSize = 13.sp, fontWeight = FontWeight.SemiBold, modifier = Modifier.clickable { onScanAgain() })
        }
    }
}
