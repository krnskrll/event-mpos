package com.eventmpos.mpos.ui.components

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.CloudOff
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.eventmpos.mpos.ui.theme.Brand
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

/**
 * Slides in a clear "offline" bar whenever the device loses internet, and slides away the moment it
 * reconnects. Selling continues offline; only order submission waits for the network.
 */
@Composable
fun OfflineBanner(isOnline: Boolean, modifier: Modifier = Modifier) {
    AnimatedVisibility(
        visible = !isOnline,
        enter = expandVertically() + fadeIn(),
        exit = shrinkVertically() + fadeOut(),
        modifier = modifier,
    ) {
        Row(
            Modifier.fillMaxWidth().background(Brand.Warn).padding(horizontal = 16.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center,
        ) {
            Icon(Icons.Outlined.CloudOff, contentDescription = null, tint = Color.White, modifier = Modifier.size(16.dp))
            Text(
                "  Offline — you can keep selling; orders submit when reconnected",
                color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold,
            )
        }
    }
}

/** Exposes app-wide connectivity to Compose (the nav shell observes this for the banner). */
@HiltViewModel
class ConnectivityViewModel @Inject constructor(
    private val syncCoordinator: com.eventmpos.mpos.core.sync.SyncCoordinator,
) : androidx.lifecycle.ViewModel() {
    val isOnline: StateFlow<Boolean> = syncCoordinator.isOnline
}
