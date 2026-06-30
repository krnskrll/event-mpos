package com.eventmpos.mpos.ui.promotions

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

/**
 * Event promotions cached on the device. These auto-apply to the cart in EPIC-07 (the BA does
 * nothing); the server confirms the authoritative discount at submit. This screen just shows what's
 * active so a BA can answer customer questions.
 */
@Composable
fun PromotionsScreen(viewModel: PromotionsViewModel = hiltViewModel()) {
    val promos by viewModel.promotions.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(20.dp)) {
        Text("Promotions", style = MaterialTheme.typography.titleLarge)
        Text(
            "Auto-applied at checkout — best one per order.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(bottom = 12.dp),
        )
        if (promos.isEmpty()) {
            Text("No promotions for this event.", color = MaterialTheme.colorScheme.onSurfaceVariant)
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                items(promos, key = { it.id }) { promo ->
                    Card(modifier = Modifier.fillMaxWidth().padding(vertical = 6.dp)) {
                        Column(Modifier.padding(16.dp)) {
                            Text(promo.name, fontWeight = FontWeight.SemiBold)
                            Text(
                                promo.describe(),
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.padding(top = 4.dp),
                            )
                        }
                    }
                }
            }
        }
    }
}
