package com.eventmpos.mpos.domain.promotion

import java.math.BigDecimal
import java.math.RoundingMode

/** A cart line for promotion preview. */
data class CartLine(val productId: String, val unitPrice: Double, val qty: Int)

data class AppliedPromotion(val promotion: Promotion, val discount: Double)

/**
 * CLIENT MIRROR of the backend engine (backend-go/internal/promotion/engine.go).
 * Computes the same client-side preview the server will authoritatively confirm at order submit.
 * **Keep this equivalent to the server logic** — if one changes, change both. Both engines run the
 * shared fixture testdata/promotion-cases.json in CI (PromotionParityTest / parity_test.go).
 *
 * The server result is authoritative; this only powers the offline cart preview (EPIC-07).
 */
object PromotionEngine {

    /** Largest discount wins; ties go to the higher priority, then the later name (same as the server). */
    fun bestPromotion(lines: List<CartLine>, promotions: List<Promotion>, nowEpochMs: Long): AppliedPromotion? =
        promotions
            .mapNotNull { p -> discountFor(p, lines, nowEpochMs)?.let { AppliedPromotion(p, it) } }
            .filter { it.discount > 0.0 }
            .maxWithOrNull(compareBy({ it.discount }, { it.promotion.priority }, { it.promotion.name }))

    fun discountFor(promo: Promotion, lines: List<CartLine>, nowEpochMs: Long): Double? {
        promo.startsAtEpochMs?.let { if (nowEpochMs < it) return null }
        promo.endsAtEpochMs?.let { if (nowEpochMs > it) return null }

        val applicable = when (promo.scope) {
            PromotionScope.ALL -> lines
            PromotionScope.SPECIFIC -> lines.filter { it.productId in promo.productIds }
        }
        if (applicable.isEmpty()) return null

        val qty = applicable.sumOf { it.qty }
        // BigDecimal.valueOf (not the BigDecimal(Double) constructor): 1.15 must stay 1.15, not
        // 1.149999…, or half-cent rounding drifts from the server's decimal math.
        val subtotal = applicable.fold(BigDecimal.ZERO) { acc, l -> acc + BigDecimal.valueOf(l.unitPrice).multiply(BigDecimal(l.qty)) }

        promo.minQty?.let { if (qty < it && promo.type != PromotionType.BUNDLE) return null }
        promo.minSubtotal?.let { if (subtotal < BigDecimal.valueOf(it)) return null }

        val raw = when (promo.type) {
            PromotionType.PERCENTAGE -> {
                val pct = promo.percentOff ?: return null
                subtotal.multiply(BigDecimal.valueOf(pct)).divide(BigDecimal(100))
            }
            PromotionType.FIXED -> BigDecimal.valueOf(promo.amountOff ?: return null)
            PromotionType.BUNDLE -> {
                val size = promo.minQty ?: return null
                if (qty < size) return null
                BigDecimal.valueOf(promo.amountOff ?: return null)
            }
        }
        val capped = raw.min(subtotal).max(BigDecimal.ZERO)
        return capped.setScale(2, RoundingMode.HALF_UP).toDouble()
    }
}
