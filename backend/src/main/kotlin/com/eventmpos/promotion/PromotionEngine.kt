package com.eventmpos.promotion

import java.math.BigDecimal
import java.math.RoundingMode
import java.time.Instant
import java.util.UUID

/** A normalized cart line for promotion evaluation. */
data class CartLine(val productId: UUID, val unitPrice: BigDecimal, val qty: Int)

/** The single promotion chosen for a cart, plus the discount it yields. */
data class AppliedPromotion(val promotion: Promotion, val discount: BigDecimal)

/**
 * Pure, deterministic promotion logic — NO framework/DB dependencies — so the Android client can
 * mirror it exactly and the server stays authoritative. **If you change a rule here, change the
 * Kotlin copy in the Android app (PromotionEngine) identically.**
 *
 * Phase 1: evaluates every active+in-window promotion, computes its discount, and returns the one
 * with the largest discount (ties broken by higher priority, then name) — one promotion per order.
 */
object PromotionEngine {

    fun bestPromotion(lines: List<CartLine>, promotions: List<Promotion>, now: Instant): AppliedPromotion? =
        promotions
            .mapNotNull { promo -> discountFor(promo, lines, now)?.let { AppliedPromotion(promo, it) } }
            .filter { it.discount > BigDecimal.ZERO }
            .maxWithOrNull(
                compareBy<AppliedPromotion>({ it.discount }, { it.promotion.priority }, { it.promotion.name }),
            )

    /** Discount this promotion yields for the cart, or null if it doesn't apply. */
    fun discountFor(promo: Promotion, lines: List<CartLine>, now: Instant): BigDecimal? {
        if (!promo.active) return null
        if (promo.startsAt != null && now.isBefore(promo.startsAt)) return null
        if (promo.endsAt != null && now.isAfter(promo.endsAt)) return null

        val applicable = when (promo.scope) {
            PromotionScope.ALL -> lines
            PromotionScope.SPECIFIC -> lines.filter { it.productId in promo.productIds }
        }
        if (applicable.isEmpty()) return null

        val qty = applicable.sumOf { it.qty }
        val subtotal = applicable.fold(BigDecimal.ZERO) { acc, l -> acc + l.unitPrice.multiply(BigDecimal(l.qty)) }

        promo.minQty?.let { if (qty < it) return null }
        promo.minSubtotal?.let { if (subtotal < it) return null }

        val raw = when (promo.type) {
            PromotionType.PERCENTAGE -> {
                val pct = promo.percentOff ?: return null
                subtotal.multiply(pct).divide(BigDecimal(100))
            }
            PromotionType.FIXED -> promo.amountOff ?: return null
            PromotionType.BUNDLE -> {
                // bundle requires its minQty (bundle size); flat amount off
                val size = promo.minQty ?: return null
                if (qty < size) return null
                promo.amountOff ?: return null
            }
        }
        // never discount more than the applicable subtotal
        val capped = raw.min(subtotal).max(BigDecimal.ZERO)
        return capped.setScale(2, RoundingMode.HALF_UP)
    }
}
