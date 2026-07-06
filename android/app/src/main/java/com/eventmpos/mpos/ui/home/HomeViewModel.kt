package com.eventmpos.mpos.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.core.sync.SyncCoordinator
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.order.OrderState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

data class HomeDashboard(
    val confirmed: Int = 0,
    val pending: Int = 0,
    val failed: Int = 0,
    val orderCount: Int = 0,
    val revenue: Double = 0.0,
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    orderRepository: OrderRepository,
    private val authRepository: AuthRepository,
    private val sessionStore: SessionStore,
    private val syncCoordinator: SyncCoordinator,
) : ViewModel() {

    val userName: String get() = sessionStore.userName.ifBlank { "Beauty Advisor" }
    val event: StateFlow<AssignedEvent?> = sessionStore.event

    private val _loggedOut = MutableStateFlow(false)
    val loggedOut: StateFlow<Boolean> = _loggedOut.asStateFlow()

    val dashboard: StateFlow<HomeDashboard> =
        orderRepository.observeOrders(sessionStore.eventId.orEmpty())
            .map { orders -> summarize(orders) }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), HomeDashboard())

    private fun summarize(orders: List<OrderRecord>) = HomeDashboard(
        confirmed = orders.count { it.state == OrderState.CONFIRMED },
        pending = orders.count { it.state == OrderState.PENDING },
        failed = orders.count { it.state == OrderState.FAILED },
        orderCount = orders.size,
        revenue = orders.filter { it.state == OrderState.CONFIRMED }.sumOf { it.total ?: 0.0 },
    )

    fun logout() {
        viewModelScope.launch {
            syncCoordinator.stop()
            authRepository.logout()
            sessionStore.clear()
            _loggedOut.value = true
        }
    }
}
