package com.eventmpos.mpos.ui.catalogue

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

/** Outcome of resolving a scanned/typed code against the event catalogue. */
sealed interface BarcodeLookup {
    data class Found(val product: Product) : BarcodeLookup
    data class NotFound(val code: String) : BarcodeLookup
}

@HiltViewModel
class BarcodeScanViewModel @Inject constructor(
    private val productRepository: ProductRepository,
    private val cartStore: CartStore,
    sessionStore: SessionStore,
) : ViewModel() {

    private val eventId: String = sessionStore.eventId.orEmpty()

    val cartCount: StateFlow<Int> =
        cartStore.items
            .map { items -> items.values.sum() }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    /** In-stock products for the emulator "simulate scan" fallback (no live camera). */
    val inStock: StateFlow<List<Product>> =
        productRepository.observeProducts(eventId)
            .map { list -> list.filter { !it.outOfStock } }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    private val _lookup = MutableStateFlow<BarcodeLookup?>(null)
    val lookup: StateFlow<BarcodeLookup?> = _lookup.asStateFlow()

    /** Resolve a code to a product (barcode match, then SKU/name) without adding to the cart yet. */
    suspend fun resolve(code: String) {
        val trimmed = code.trim()
        if (trimmed.isEmpty()) return
        val byBarcode = productRepository.findByBarcode(eventId, trimmed)
        val product = byBarcode ?: inStock.value.firstOrNull {
            it.sku.equals(trimmed, true) || it.name.contains(trimmed, true)
        }
        _lookup.value = if (product != null) BarcodeLookup.Found(product) else BarcodeLookup.NotFound(trimmed)
    }

    /** Show a specific product in the result sheet (used by the simulate-scan fallback). */
    fun show(product: Product) { _lookup.value = BarcodeLookup.Found(product) }

    fun addToCart(product: Product, qty: Int = 1) {
        if (!product.outOfStock) cartStore.add(product.id, qty)
    }

    fun dismiss() { _lookup.value = null }
}
