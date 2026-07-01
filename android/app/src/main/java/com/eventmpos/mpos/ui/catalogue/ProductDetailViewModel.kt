package com.eventmpos.mpos.ui.catalogue

import androidx.lifecycle.SavedStateHandle
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
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ProductDetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val productRepository: ProductRepository,
    private val cartStore: CartStore,
    sessionStore: SessionStore,
) : ViewModel() {

    private val productId: String = checkNotNull(savedStateHandle["productId"])
    private val eventId: String = sessionStore.eventId.orEmpty()

    private val _product = MutableStateFlow<Product?>(null)
    val product: StateFlow<Product?> = _product.asStateFlow()

    val cartCount: StateFlow<Int> = cartStore.items
        .map { it.values.sum() }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    init {
        viewModelScope.launch { _product.value = productRepository.getProduct(eventId, productId) }
    }

    fun addToCart(qty: Int) {
        val p = _product.value ?: return
        if (!p.outOfStock && qty > 0) cartStore.add(p.id, qty)
    }
}
