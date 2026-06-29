package com.eventmpos.mpos.ui.catalogue

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.core.scanner.ScannerPort
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
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
    savedStateHandle: SavedStateHandle,
    private val productRepository: ProductRepository,
    private val scanner: ScannerPort,
) : ViewModel() {

    private val eventId: String = checkNotNull(savedStateHandle["eventId"])

    private val query = MutableStateFlow("")
    val searchQuery: StateFlow<String> = query.asStateFlow()

    private val _scan = MutableStateFlow<ScanFeedback?>(null)
    val scan: StateFlow<ScanFeedback?> = _scan.asStateFlow()

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
        // Hardware scans (Sunmi) flow in here; resolve each against the local catalogue.
        viewModelScope.launch {
            scanner.scans.collect { onBarcode(it) }
        }
    }

    fun onSearch(value: String) { query.value = value }

    /** Manual/typed barcode entry — routed through the same scanner port as hardware scans. */
    fun submitManualBarcode(code: String) {
        if (code.isNotBlank()) viewModelScope.launch { scanner.emitManual(code) }
    }

    fun clearScan() { _scan.value = null }

    private suspend fun onBarcode(code: String) {
        val product = productRepository.findByBarcode(eventId, code)
        _scan.value = if (product != null) ScanFeedback.Found(product) else ScanFeedback.NotFound(code)
    }
}
