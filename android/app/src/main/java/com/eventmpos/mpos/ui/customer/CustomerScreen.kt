package com.eventmpos.mpos.ui.customer

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
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
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
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.components.SectionCard
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun CustomerScreen(
    onBack: () -> Unit,
    onDone: () -> Unit,
    viewModel: CustomerViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var qr by remember { mutableStateOf("") }
    var manualName by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize().background(Brand.Bg)) {
        Column(Modifier.fillMaxWidth().background(Brand.Navy).padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(32.dp).clip(CircleShape).background(Color.White.copy(alpha = 0.12f)).clickable { onBack() }, Alignment.Center) {
                    Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back", tint = Color.White, modifier = Modifier.size(18.dp))
                }
                Spacer(Modifier.size(12.dp))
                Text("Identify customer", color = Color.White, fontSize = 18.sp, fontWeight = FontWeight.Bold)
            }
        }

        Column(Modifier.fillMaxSize().padding(20.dp)) {
            Text("Scan or type the customer's QR", color = Brand.TextMuted, fontSize = 13.sp, modifier = Modifier.padding(bottom = 8.dp))
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = qr, onValueChange = { qr = it }, singleLine = true,
                    placeholder = { Text("e.g. MBR-00001", color = Brand.TextFaint) },
                    shape = RoundedCornerShape(12.dp), modifier = Modifier.weight(1f),
                )
                BrandButton("Identify", onClick = { viewModel.identifyByQr(qr) }, enabled = !state.identifying, modifier = Modifier.size(width = 110.dp, height = 56.dp))
            }

            if (state.identifying) {
                Spacer(Modifier.size(16.dp)); CircularProgressIndicator(color = Brand.Navy)
            }

            state.selected?.let { c ->
                SectionCard(modifier = Modifier.padding(top = 16.dp)) {
                    Text("CUSTOMER FOR THIS ORDER", color = Brand.TextFaint, fontSize = 11.sp, letterSpacing = 1.sp)
                    Text(c.name, color = Brand.Heading, fontSize = 18.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(top = 4.dp))
                    Text(
                        "${c.tier ?: "—"} · " + when (c.source) {
                            CustomerSource.LOCAL_CACHE -> "matched offline"
                            CustomerSource.ONLINE_LOOKUP -> "matched online"
                            CustomerSource.MANUAL -> "manually entered"
                        },
                        color = Brand.TextMuted, fontSize = 13.sp, modifier = Modifier.padding(top = 2.dp),
                    )
                }
                BrandButton("Continue", onClick = onDone, color = Brand.Teal, modifier = Modifier.fillMaxWidth().padding(top = 16.dp))
            }

            if (state.notFoundQr != null && state.selected == null) {
                Spacer(Modifier.size(16.dp))
                Text("No customer found for \"${state.notFoundQr}\". Enter a name to proceed:", color = Brand.Danger, fontSize = 13.sp)
                OutlinedTextField(
                    value = manualName, onValueChange = { manualName = it }, singleLine = true,
                    placeholder = { Text("Customer name") }, shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                )
                BrandButton("Use this name", onClick = { viewModel.setManual(manualName) }, modifier = Modifier.fillMaxWidth().padding(top = 8.dp))
            }
        }
    }
}
