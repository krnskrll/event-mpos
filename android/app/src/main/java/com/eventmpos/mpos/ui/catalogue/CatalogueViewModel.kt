package com.eventmpos.mpos.ui.catalogue

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.core.scanner.ScannerPort
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import javax.inject.Inject

/** Result of a scan: the matched product, or a "not found" for an unknown barcode. */
sealed interface ScanFeedback {
    data class Found(val product: Product) : ScanFeedback
    data class NotFound(val barcode: String) : ScanFeedback
}

@HiltViewModel
class CatalogueViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val productRepository: ProductRepository,
    private val scanner: ScannerPort,
    private val cartStore: CartStore,
) : ViewModel() {

    val eventId: String = checkNotNull(savedStateHandle["eventId"])

    private val query = MutableStateFlow("")
    val searchQuery: StateFlow<String> = query.asStateFlow()

    private val _scan = MutableStateFlow<ScanFeedback?>(null)
    val scan: StateFlow<ScanFeedback?> = _scan.asStateFlow()

    /** Total item count in the cart, for the "View cart (N)" affordance. */
    val cartCount: StateFlow<Int> =
        cartStore.items
            .map { items -> items.values.sum() }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    /** Products for this event from Room, filtered by the search query. Offline-friendly. */
    val products: StateFlow<List<Product>> =
        productRepository.observeProducts(eventId)
            .combine(query) { list, q ->
                if (q.isBlank()) list
                else list.filter {
                    it.name.contains(q, true) || it.sku.contains(q, true) ||
                        it.barcodes.any { bc -> bc.contains(q, true) }
                }
            }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    init {
        viewModelScope.launch { scanner.scans.collect { onBarcode(it) } }
        viewModelScope.launch {
            while (isActive) {
                productRepository.refreshStock(eventId)
                delay(10_000)
            }
        }
    }

    fun onSearch(value: String) { query.value = value }

    fun addToCart(product: Product) {
        if (!product.outOfStock) cartStore.add(product.id)
    }

    fun submitManualBarcode(code: String) {
        if (code.isNotBlank()) viewModelScope.launch { scanner.emitManual(code) }
    }

    fun clearScan() { _scan.value = null }

    private suspend fun onBarcode(code: String) {
        val product = productRepository.findByBarcode(eventId, code)
        if (product != null && !product.outOfStock) cartStore.add(product.id)
        _scan.value = if (product != null) ScanFeedback.Found(product) else ScanFeedback.NotFound(code)
    }
}
