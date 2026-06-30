package com.eventmpos.mpos.data.cart

import com.eventmpos.mpos.domain.member.SelectedCustomer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Entirely-local cart state (per .cursorrules: no server calls while browsing/adjusting). Holds
 * productId→qty and the selected customer for the current sale. Survives navigation but not process
 * death — a deliberate "draft to resume" is persisted separately (see CartRepository.saveDraft).
 */
@Singleton
class CartStore @Inject constructor() {
    private val _items = MutableStateFlow<Map<String, Int>>(emptyMap())
    val items: StateFlow<Map<String, Int>> = _items.asStateFlow()

    private val _customer = MutableStateFlow<SelectedCustomer?>(null)
    val customer: StateFlow<SelectedCustomer?> = _customer.asStateFlow()

    fun add(productId: String, delta: Int = 1) = _items.update { current ->
        val next = (current[productId] ?: 0) + delta
        if (next <= 0) current - productId else current + (productId to next)
    }

    fun setQty(productId: String, qty: Int) = _items.update { current ->
        if (qty <= 0) current - productId else current + (productId to qty)
    }

    fun remove(productId: String) = _items.update { it - productId }

    fun setCustomer(c: SelectedCustomer?) { _customer.value = c }

    fun clear() {
        _items.value = emptyMap()
        _customer.value = null
    }

    fun isEmpty() = _items.value.isEmpty()
}
