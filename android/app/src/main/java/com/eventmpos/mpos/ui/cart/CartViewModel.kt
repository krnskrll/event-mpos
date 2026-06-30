package com.eventmpos.mpos.ui.cart

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.domain.member.SelectedCustomer
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import com.eventmpos.mpos.domain.promotion.CartLine
import com.eventmpos.mpos.domain.promotion.PromotionEngine
import com.eventmpos.mpos.domain.promotion.PromotionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import java.math.BigDecimal
import java.math.RoundingMode
import javax.inject.Inject

data class CartLineItem(val product: Product, val qty: Int) {
    val lineTotal: Double get() = product.price * qty
}

data class CartUiState(
    val lines: List<CartLineItem> = emptyList(),
    val customer: SelectedCustomer? = null,
    val subtotal: Double = 0.0,
    val discount: Double = 0.0,
    val total: Double = 0.0,
    val promotionName: String? = null,
    val submitting: Boolean = false,
    val submitted: Boolean = false,
)

@HiltViewModel
class CartViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val cartStore: CartStore,
    private val productRepository: ProductRepository,
    private val promotionRepository: PromotionRepository,
    private val orderRepository: OrderRepository,
) : ViewModel() {

    private val eventId: String = checkNotNull(savedStateHandle["eventId"])
    private val submitting = MutableStateFlow(false)
    private val submitted = MutableStateFlow(false)

    val state: StateFlow<CartUiState> =
        combine(
            cartStore.items,
            productRepository.observeProducts(eventId),
            promotionRepository.observePromotions(eventId),
            cartStore.customer,
            submitting,
        ) { items, products, promos, customer, isSubmitting ->
            val byId = products.associateBy { it.id }
            val lines = items.mapNotNull { (id, qty) -> byId[id]?.let { CartLineItem(it, qty) } }
            val subtotal = lines.fold(BigDecimal.ZERO) { acc, l -> acc + BigDecimal(l.product.price).multiply(BigDecimal(l.qty)) }
            // local promo preview (server confirms authoritatively at submit)
            val cartLines = lines.map { CartLine(it.product.id, it.product.price, it.qty) }
            val applied = PromotionEngine.bestPromotion(cartLines, promos, System.currentTimeMillis())
            val discount = applied?.discount ?: 0.0
            val total = (subtotal.toDouble() - discount).coerceAtLeast(0.0)
            CartUiState(
                lines = lines,
                customer = customer,
                subtotal = subtotal.setScale(2, RoundingMode.HALF_UP).toDouble(),
                discount = discount,
                total = BigDecimal(total).setScale(2, RoundingMode.HALF_UP).toDouble(),
                promotionName = applied?.promotion?.name,
                submitting = isSubmitting,
                submitted = submitted.value,
            )
        }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), CartUiState())

    fun increment(productId: String) = cartStore.add(productId, 1)
    fun decrement(productId: String) = cartStore.add(productId, -1)
    fun remove(productId: String) = cartStore.remove(productId)

    val canConfirm: Boolean
        get() = !cartStore.isEmpty() && cartStore.customer.value != null

    fun confirm(onDone: () -> Unit) {
        if (!canConfirm || submitting.value) return
        submitting.value = true
        viewModelScope.launch {
            runCatching { orderRepository.confirmCurrentCart(eventId) }
            submitting.value = false
            submitted.value = true
            onDone()
        }
    }
}
