package com.eventmpos.mpos.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

private val LightColors = lightColorScheme(
    primary = Brand.Navy,
    onPrimary = androidx.compose.ui.graphics.Color.White,
    secondary = Brand.Teal,
    onSecondary = androidx.compose.ui.graphics.Color.White,
    background = Brand.Bg,
    onBackground = Brand.Text,
    surface = Brand.Card,
    onSurface = Brand.Text,
    surfaceVariant = Brand.CardBorder,
    onSurfaceVariant = Brand.TextMuted,
    error = Brand.Danger,
)

/** Light-only theme matching the Phase-1 design (navy + teal). */
@Composable
fun EventMposTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = LightColors, content = content)
}
