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
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.domain.auth.EventStatus
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.flatMapLatest
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.flow.stateIn
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
 * reconnects. Selling continues offline (EPIC-12 C1); only the upload waits for the network.
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
                "  Offline — keep selling; orders get final numbers and print now, and sync when reconnected",
                color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold,
            )
        }
    }
}

/** Exposes app-wide connectivity (and the closed-event sync state) to the nav shell's banners. */
@HiltViewModel
class ConnectivityViewModel @Inject constructor(
    syncCoordinator: com.eventmpos.mpos.core.sync.SyncCoordinator,
    session: com.eventmpos.mpos.data.session.SessionStore,
    pendingOrders: com.eventmpos.mpos.data.local.PendingOrderDao,
) : androidx.lifecycle.ViewModel() {
    val isOnline: StateFlow<Boolean> = syncCoordinator.isOnline

    /** Orders still waiting to upload after the event closed (0 while the event is open). */
    @OptIn(kotlinx.coroutines.ExperimentalCoroutinesApi::class)
    val closedEventPending: StateFlow<Int> = session.event
        .flatMapLatest { e -> if (e?.status == EventStatus.CLOSED) pendingOrders.observePendingCount(e.id) else flowOf(0) }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)
}

/** The event closed while this device still has orders to upload: keep the app open and online. */
@Composable
fun ClosedEventBanner(pending: Int) {
    AnimatedVisibility(visible = pending > 0, enter = expandVertically() + fadeIn(), exit = shrinkVertically() + fadeOut()) {
        Text(
            "Event closed — $pending order(s) still syncing. Keep the app open and online; they're accepted for 24h.",
            color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold,
            modifier = Modifier.fillMaxWidth().background(Brand.Navy).padding(horizontal = 16.dp, vertical = 8.dp),
        )
    }
}
