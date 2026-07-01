package com.eventmpos.mpos.ui.catalogue

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.core.connectivity.ConnectivityObserver
import com.eventmpos.mpos.core.scanner.ScannerPort
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.event.EventRepository
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

/** Result of a scan: the matched product, or a "not found" for an unknown barcode. */
sealed interface ScanFeedback {
    data class Found(val product: Product) : ScanFeedback
    data class NotFound(val barcode: String) : ScanFeedback
}

@HiltViewModel
class CatalogueViewModel @Inject constructor(
    private val productRepository: ProductRepository,
    private val eventRepository: EventRepository,
    private val scanner: ScannerPort,
    private val cartStore: CartStore,
    connectivityObserver: ConnectivityObserver,
    sessionStore: SessionStore,
) : ViewModel() {

    val eventId: String = sessionStore.eventId.orEmpty()
    val eventName: String = sessionStore.event.value?.name ?: "Current event"

    /** Live connectivity for the header indicator (Synced vs Offline). */
    val isOnline: StateFlow<Boolean> = connectivityObserver.isOnline

    /** True while a manual pull-to-refresh (full re-sync) is running. */
    private val _refreshing = MutableStateFlow(false)
    val refreshing: StateFlow<Boolean> = _refreshing.asStateFlow()

    private val query = MutableStateFlow("")
    val searchQuery: StateFlow<String> = query.asStateFlow()

    private val _category = MutableStateFlow("All")
    val category: StateFlow<String> = _category.asStateFlow()

    private val _scan = MutableStateFlow<ScanFeedback?>(null)
    val scan: StateFlow<ScanFeedback?> = _scan.asStateFlow()

    /** Active customer linked to the sale (shown in the catalogue header). */
    val customer: StateFlow<com.eventmpos.mpos.domain.member.SelectedCustomer?> =
        cartStore.customer.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), null)

    /** Category chips = "All" + distinct brands present in the catalogue. */
    val categories: StateFlow<List<String>> =
        productRepository.observeProducts(eventId)
            .map { list -> listOf("All") + list.mapNotNull { it.brand }.distinct().sorted() }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), listOf("All"))

    /** Total item count in the cart, for the "View cart (N)" affordance. */
    val cartCount: StateFlow<Int> =
        cartStore.items
            .map { items -> items.values.sum() }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    /** Products for this event from Room, filtered by the search query. Offline-friendly. */
    val products: StateFlow<List<Product>> =
        combine(productRepository.observeProducts(eventId), query, _category) { list, q, cat ->
            list.filter { p ->
                (cat == "All" || p.brand == cat) &&
                    (q.isBlank() || p.name.contains(q, true) || p.sku.contains(q, true) ||
                        p.barcodes.any { bc -> bc.contains(q, true) })
            }
        }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    init {
        // Hardware-scanner emissions still resolve here; periodic stock refresh now runs in the
        // app-scoped SyncCoordinator (every 60s) so it survives leaving this screen.
        viewModelScope.launch { scanner.scans.collect { onBarcode(it) } }
    }

    fun onSearch(value: String) { query.value = value }
    fun onCategory(value: String) { _category.value = value }

    /** Pull-to-refresh: full event re-sync (products/members/promotions) + fresh stock. */
    fun refresh() {
        if (_refreshing.value) return
        _refreshing.value = true
        viewModelScope.launch {
            runCatching {
                eventRepository.syncSnapshot(eventId)
                productRepository.refreshStock(eventId)
            }
            _refreshing.value = false
        }
    }

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
