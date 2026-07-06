package com.eventmpos.mpos.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.eventmpos.mpos.ui.theme.Brand

/** Primary filled action button (navy by default; pass Brand.Teal for the accent variant). */
@Composable
fun BrandButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    loading: Boolean = false,
    color: Color = Brand.Navy,
) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(12.dp))
            .background(if (enabled) color else Brand.FieldBorder)
            .clickable(enabled = enabled && !loading, onClick = onClick)
            .height(52.dp),
        contentAlignment = Alignment.Center,
    ) {
        if (loading) {
            CircularProgressIndicator(color = Color.White, strokeWidth = 2.dp, modifier = Modifier.height(20.dp))
        } else {
            Text(text, color = if (enabled) Color.White else Brand.TextFaint, fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
        }
    }
}

/** Rounded status/label chip. */
@Composable
fun Pill(text: String, color: Color, bg: Color, modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.clip(RoundedCornerShape(999.dp)).background(bg).padding(horizontal = 10.dp, vertical = 4.dp),
    ) {
        Text(text, color = color, fontSize = 11.sp, fontWeight = FontWeight.Bold)
    }
}

/** White rounded card with a subtle border. */
@Composable
fun SectionCard(
    modifier: Modifier = Modifier,
    padding: PaddingValues = PaddingValues(16.dp),
    content: @Composable ColumnScope.() -> Unit,
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(16.dp))
            .background(Brand.Card)
            .padding(padding),
        content = content,
    )
}

/** Navy header block with status-bar top padding. */
@Composable
fun NavyHeader(content: @Composable ColumnScope.() -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Brand.Navy)
            .padding(top = 44.dp, start = 20.dp, end = 20.dp, bottom = 20.dp),
        content = content,
    )
}

typealias ColumnScope = androidx.compose.foundation.layout.ColumnScope
