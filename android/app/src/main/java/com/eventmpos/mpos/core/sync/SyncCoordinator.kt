package com.eventmpos.mpos.core.sync

import com.eventmpos.mpos.core.connectivity.ConnectivityObserver
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.cart.CartRepository
import com.eventmpos.mpos.data.order.OrderNumbers
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.product.ProductRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

/**
 * App-scoped background sync for an active selling session. Runs independently of any screen so the
 * BA's current view is never disrupted:
 *  - refreshes stock silently every 60s (Room is updated; screens observing it just update)
 *  - on a offline→online transition, flushes the order retry queue and refreshes stock immediately
 *  - keeps the device's order-number block topped up, and notices when the event closes
 *
 * Started when the event snapshot finishes syncing and stopped on logout / session end. Restarting
 * with a new event cancels the previous session's loops.
 */
@Singleton
class SyncCoordinator @Inject constructor(
    private val productRepository: ProductRepository,
    private val orderRepository: OrderRepository,
    private val numbers: OrderNumbers,
    private val api: AuthApi,
    private val session: SessionStore,
    private val carts: CartRepository,
    connectivity: ConnectivityObserver,
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private var sessionJob: Job? = null

    val isOnline: StateFlow<Boolean> = connectivity.isOnline

    fun start(eventId: String) {
        sessionJob?.cancel()
        sessionJob = scope.launch {
            // Silent periodic stock refresh (best-effort; offline = no-op, keeps cached levels).
            launch {
                while (isActive) {
                    productRepository.refreshStock(eventId)
                    numbers.topUp(eventId)
                    refreshEventStatus(eventId)
                    delay(STOCK_REFRESH_MS)
                }
            }
            // Reconnect trigger: one signal drives both the retry-queue flush and a stock refresh.
            launch {
                var wasOnline = isOnline.value
                isOnline.collect { online ->
                    if (online && !wasOnline) {
                        orderRepository.retryPending()
                        productRepository.refreshStock(eventId)
                        numbers.topUp(eventId)
                    }
                    wasOnline = online
                }
            }
        }
    }

    /**
     * Picks up an event closing mid-session (D2): parked carts are dropped, and queued orders keep
     * uploading — the server accepts orders created before the close for 24h.
     */
    private suspend fun refreshEventStatus(eventId: String) {
        val status = runCatching { api.getMyEvents().firstOrNull { it.id == eventId }?.status }.getOrNull() ?: return
        val parsed = runCatching { EventStatus.valueOf(status) }.getOrNull() ?: return
        if (parsed == session.event.value?.status) return
        session.updateStatus(parsed)
        if (parsed == EventStatus.CLOSED) carts.clearEvent(eventId)
    }

    fun stop() {
        sessionJob?.cancel()
        sessionJob = null
    }

    private companion object {
        const val STOCK_REFRESH_MS = 60_000L
    }
}
