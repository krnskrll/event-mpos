package com.eventmpos.mpos.ui.orders

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class OrdersViewModel @Inject constructor(
    private val orderRepository: OrderRepository,
    sessionStore: SessionStore,
) : ViewModel() {

    private val eventId: String = sessionStore.eventId.orEmpty()

    val orders: StateFlow<List<OrderRecord>> =
        orderRepository.observeOrders(eventId)
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    /** Result of the last cancel attempt, shown above the list. */
    val message = MutableStateFlow<String?>(null)

    fun cancel(orderKey: String) = viewModelScope.launch {
        message.value = orderRepository.cancel(orderKey) ?: "Order cancelled — stock returned."
    }
}
