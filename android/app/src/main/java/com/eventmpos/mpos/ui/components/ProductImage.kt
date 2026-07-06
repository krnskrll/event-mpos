package com.eventmpos.mpos.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import coil.compose.AsyncImage
import com.eventmpos.mpos.ui.theme.Brand

/**
 * Product image with a branded placeholder. Product photos aren't always available/representative,
 * so when the URL is null/blank we render a neutral light-indigo tile with an image glyph instead of
 * a broken/empty image. Callers supply sizing/clipping via [modifier].
 */
@Composable
fun ProductImage(imageUrl: String?, contentDescription: String?, modifier: Modifier = Modifier) {
    if (!imageUrl.isNullOrBlank()) {
        AsyncImage(
            model = imageUrl,
            contentDescription = contentDescription,
            contentScale = ContentScale.Crop,
            modifier = modifier,
        )
    } else {
        Box(modifier.background(Brand.NavActiveBg), contentAlignment = Alignment.Center) {
            Icon(
                Icons.Outlined.Image,
                contentDescription = contentDescription,
                tint = Brand.TextGhost,
                modifier = Modifier.fillMaxSize(0.32f),
            )
        }
    }
}
