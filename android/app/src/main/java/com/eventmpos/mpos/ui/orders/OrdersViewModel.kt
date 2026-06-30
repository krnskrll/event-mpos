package com.eventmpos.mpos.ui.orders

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

@HiltViewModel
class OrdersViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    orderRepository: OrderRepository,
) : ViewModel() {

    private val eventId: String = checkNotNull(savedStateHandle["eventId"])

    val orders: StateFlow<List<OrderRecord>> =
        orderRepository.observeOrders(eventId)
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())
}
