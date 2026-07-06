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
/** Encodes a cart line's product + optional variant into a single map key: "productId|variantId". */
object CartKey {
    fun of(productId: String, variantId: String?): String = productId + "|" + (variantId ?: "")
    fun productId(key: String): String = key.substringBefore("|")
    fun variantId(key: String): String? = key.substringAfter("|", "").ifEmpty { null }
}

@Singleton
class CartStore @Inject constructor() {
    // Keyed by CartKey (product + optional variant) so different shades of a product are separate lines.
    private val _items = MutableStateFlow<Map<String, Int>>(emptyMap())
    val items: StateFlow<Map<String, Int>> = _items.asStateFlow()

    private val _customer = MutableStateFlow<SelectedCustomer?>(null)
    val customer: StateFlow<SelectedCustomer?> = _customer.asStateFlow()

    /** Add a base product (no variant). Existing catalogue/scan callers use this. */
    fun add(productId: String, delta: Int = 1) = adjust(CartKey.of(productId, null), delta)

    /** Add a specific variant (from the product detail picker). */
    fun addVariant(productId: String, variantId: String?, delta: Int = 1) =
        adjust(CartKey.of(productId, variantId), delta)

    /** Adjust a cart line by its encoded key (used by the cart's +/− steppers). */
    fun adjust(cartKey: String, delta: Int) = _items.update { current ->
        val next = (current[cartKey] ?: 0) + delta
        if (next <= 0) current - cartKey else current + (cartKey to next)
    }

    fun remove(cartKey: String) = _items.update { it - cartKey }

    fun setCustomer(c: SelectedCustomer?) { _customer.value = c }

    fun clear() {
        _items.value = emptyMap()
        _customer.value = null
    }

    fun isEmpty() = _items.value.isEmpty()
}
