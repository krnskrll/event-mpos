package com.eventmpos.mpos.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Assignment
import androidx.compose.material.icons.outlined.Home
import androidx.compose.material.icons.outlined.ShoppingBag
import androidx.compose.material.icons.outlined.ShoppingCart
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.eventmpos.mpos.ui.theme.Brand

/** Persistent bottom navigation (Home / Catalogue / Cart / Orders) matching the Phase-1 design. */
@Composable
fun BottomBar(currentRoute: String?, cartCount: Int, onSelect: (String) -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth().background(Color.White).padding(vertical = 6.dp),
        horizontalArrangement = Arrangement.SpaceAround,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        NavItem(Icons.Outlined.Home, "Home", currentRoute == Routes.HOME) { onSelect(Routes.HOME) }
        NavItem(Icons.Outlined.ShoppingBag, "Catalogue", currentRoute == Routes.CATALOGUE) { onSelect(Routes.CATALOGUE) }
        NavItem(Icons.Outlined.ShoppingCart, "Cart", currentRoute == Routes.CART, badge = cartCount) { onSelect(Routes.CART) }
        NavItem(Icons.AutoMirrored.Outlined.Assignment, "Orders", currentRoute == Routes.ORDERS) { onSelect(Routes.ORDERS) }
    }
}

@Composable
private fun NavItem(icon: ImageVector, label: String, active: Boolean, badge: Int = 0, onClick: () -> Unit) {
    val tint = if (active) Brand.Navy else Brand.TextMuted
    Column(
        modifier = Modifier.clip(RoundedCornerShape(12.dp)).background(if (active) Brand.NavActiveBg else Color.Transparent)
            .clickable { onClick() }.padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Box {
            Icon(icon, contentDescription = label, tint = tint, modifier = Modifier.size(24.dp))
            if (badge > 0) {
                Box(Modifier.align(Alignment.TopEnd).size(15.dp).clip(CircleShape).background(Brand.Teal), Alignment.Center) {
                    Text(if (badge > 9) "9+" else "$badge", color = Color.White, fontSize = 9.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
        Text(label, color = tint, fontSize = 10.sp, fontWeight = if (active) FontWeight.SemiBold else FontWeight.Normal)
    }
}
