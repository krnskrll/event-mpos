package com.eventmpos.mpos.ui.orders

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

@HiltViewModel
class OrdersViewModel @Inject constructor(
    orderRepository: OrderRepository,
    sessionStore: SessionStore,
) : ViewModel() {

    private val eventId: String = sessionStore.eventId.orEmpty()

    val orders: StateFlow<List<OrderRecord>> =
        orderRepository.observeOrders(eventId)
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())
}
