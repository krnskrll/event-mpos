package com.eventmpos.mpos.ui.cart

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.cart.CartDraft
import com.eventmpos.mpos.data.cart.CartKey
import com.eventmpos.mpos.data.cart.CartRepository
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.member.SelectedCustomer
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.order.Sale
import com.eventmpos.mpos.domain.receipt.ReceiptLine
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductVariant
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

data class CartLineItem(
    val key: String,
    val product: Product,
    val variant: ProductVariant?,
    val qty: Int,
) {
    val unitPrice: Double get() = variant?.price ?: product.price
    val displayName: String get() = variant?.let { "${product.name} — ${it.name}" } ?: product.name
    val lineTotal: Double get() = unitPrice * qty
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
    private val cartStore: CartStore,
    private val productRepository: ProductRepository,
    private val promotionRepository: PromotionRepository,
    private val orderRepository: OrderRepository,
    private val cartRepository: CartRepository,
    sessionStore: SessionStore,
) : ViewModel() {

    private val eventId: String = sessionStore.eventId.orEmpty()
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
            val lines = items.mapNotNull { (key, qty) ->
                val p = byId[CartKey.productId(key)] ?: return@mapNotNull null
                val variant = CartKey.variantId(key)?.let { vid -> p.variants.firstOrNull { it.id == vid } }
                CartLineItem(key, p, variant, qty)
            }
            val subtotal = lines.fold(BigDecimal.ZERO) { acc, l -> acc + BigDecimal.valueOf(l.unitPrice).multiply(BigDecimal(l.qty)) }
            // local promo preview (server confirms authoritatively at submit)
            val cartLines = lines.map { CartLine(it.product.id, it.unitPrice, it.qty) }
            val applied = PromotionEngine.bestPromotion(cartLines, promos, System.currentTimeMillis())
            val discount = applied?.discount ?: 0.0
            val total = (subtotal.toDouble() - discount).coerceAtLeast(0.0)
            CartUiState(
                lines = lines,
                customer = customer,
                subtotal = subtotal.setScale(2, RoundingMode.HALF_UP).toDouble(),
                discount = discount,
                total = BigDecimal.valueOf(total).setScale(2, RoundingMode.HALF_UP).toDouble(),
                promotionName = applied?.promotion?.name,
                submitting = isSubmitting,
                submitted = submitted.value,
            )
        }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), CartUiState())

    val drafts: StateFlow<List<CartDraft>> = cartRepository.observeDrafts(eventId)
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    fun saveDraft(label: String) = viewModelScope.launch { cartRepository.saveDraft(label) }
    fun resumeDraft(id: String) = viewModelScope.launch { cartRepository.resume(id) }
    fun deleteDraft(id: String) = viewModelScope.launch { cartRepository.delete(id) }

    fun increment(cartKey: String) = cartStore.adjust(cartKey, 1)
    fun decrement(cartKey: String) = cartStore.adjust(cartKey, -1)
    fun remove(cartKey: String) = cartStore.remove(cartKey)

    val canConfirm: Boolean
        get() = !cartStore.isEmpty() && cartStore.customer.value != null

    /** Why the last confirm didn't go through (e.g. offline with no order numbers left). */
    val error = MutableStateFlow<String?>(null)

    /** Confirms on the device (final number + receipt), online or offline; the upload follows. */
    fun confirm(onDone: () -> Unit) {
        if (!canConfirm || submitting.value) return
        val s = state.value
        val sale = Sale(
            items = s.lines.map { ReceiptLine(it.displayName, it.qty, it.lineTotal) },
            subtotal = s.subtotal, discount = s.discount, promotionName = s.promotionName, total = s.total,
        )
        submitting.value = true
        error.value = null
        viewModelScope.launch {
            val result = runCatching { orderRepository.confirmCurrentCart(eventId, sale) }
            submitting.value = false
            result.onSuccess { submitted.value = true; onDone() }
                .onFailure { error.value = it.message ?: "Could not confirm the order" }
        }
    }
}
