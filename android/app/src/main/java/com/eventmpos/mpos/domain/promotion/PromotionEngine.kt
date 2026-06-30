package com.eventmpos.mpos.domain.promotion

import java.math.BigDecimal
import java.math.RoundingMode

/** A cart line for promotion preview. */
data class CartLine(val productId: String, val unitPrice: Double, val qty: Int)

data class AppliedPromotion(val promotion: Promotion, val discount: Double)

/**
 * CLIENT MIRROR of the backend PromotionEngine (com.eventmpos.promotion.PromotionEngine).
 * Computes the same client-side preview the server will authoritatively confirm at order submit.
 * **Keep this byte-for-byte equivalent to the server logic** — if one changes, change both.
 *
 * The server result is authoritative; this only powers the offline cart preview (EPIC-07).
 */
object PromotionEngine {

    fun bestPromotion(lines: List<CartLine>, promotions: List<Promotion>, nowEpochMs: Long): AppliedPromotion? =
        promotions
            .mapNotNull { p -> discountFor(p, lines, nowEpochMs)?.let { AppliedPromotion(p, it) } }
            .filter { it.discount > 0.0 }
            .maxWithOrNull(compareBy({ it.discount }, { it.promotion.name }))

    fun discountFor(promo: Promotion, lines: List<CartLine>, nowEpochMs: Long): Double? {
        promo.startsAtEpochMs?.let { if (nowEpochMs < it) return null }
        promo.endsAtEpochMs?.let { if (nowEpochMs > it) return null }

        val applicable = when (promo.scope) {
            PromotionScope.ALL -> lines
            PromotionScope.SPECIFIC -> lines.filter { it.productId in promo.productIds }
        }
        if (applicable.isEmpty()) return null

        val qty = applicable.sumOf { it.qty }
        val subtotal = applicable.fold(BigDecimal.ZERO) { acc, l -> acc + BigDecimal(l.unitPrice).multiply(BigDecimal(l.qty)) }

        promo.minQty?.let { if (qty < it && promo.type != PromotionType.BUNDLE) return null }
        promo.minSubtotal?.let { if (subtotal < BigDecimal(it)) return null }

        val raw = when (promo.type) {
            PromotionType.PERCENTAGE -> {
                val pct = promo.percentOff ?: return null
                subtotal.multiply(BigDecimal(pct)).divide(BigDecimal(100))
            }
            PromotionType.FIXED -> BigDecimal(promo.amountOff ?: return null)
            PromotionType.BUNDLE -> {
                val size = promo.minQty ?: return null
                if (qty < size) return null
                BigDecimal(promo.amountOff ?: return null)
            }
        }
        val capped = raw.min(subtotal).max(BigDecimal.ZERO)
        return capped.setScale(2, RoundingMode.HALF_UP).toDouble()
    }
}
