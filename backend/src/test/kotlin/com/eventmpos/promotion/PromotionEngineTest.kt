package com.eventmpos.promotion

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.Test
import java.math.BigDecimal
import java.time.Instant
import java.time.temporal.ChronoUnit
import java.util.UUID

/** Verifies the shared promotion logic (the Android engine must match these expectations). */
class PromotionEngineTest {

    private val pA = UUID.randomUUID()
    private val pB = UUID.randomUUID()
    private val now = Instant.parse("2026-07-05T10:00:00Z")

    private fun cart() = listOf(
        CartLine(pA, BigDecimal("100.00"), 2), // 200
        CartLine(pB, BigDecimal("50.00"), 1),  // 50  -> subtotal 250
    )

    private fun promo(
        type: PromotionType, name: String = "p", pct: String? = null, amt: String? = null,
        scope: PromotionScope = PromotionScope.ALL, products: List<UUID> = emptyList(),
        minQty: Int? = null, minSub: String? = null, start: Instant? = null, end: Instant? = null,
        active: Boolean = true, priority: Int = 0,
    ) = Promotion(
        eventId = UUID.randomUUID(), name = name, type = type,
        percentOff = pct?.let { BigDecimal(it) }, amountOff = amt?.let { BigDecimal(it) },
        scope = scope, minQty = minQty, minSubtotal = minSub?.let { BigDecimal(it) },
        startsAt = start, endsAt = end, active = active, priority = priority,
    ).apply { productIds = products }

    @Test
    fun `percentage applies to whole cart`() {
        val d = PromotionEngine.discountFor(promo(PromotionType.PERCENTAGE, pct = "10"), cart(), now)
        assertEquals(BigDecimal("25.00"), d) // 10% of 250
    }

    @Test
    fun `fixed requires min subtotal`() {
        val ok = promo(PromotionType.FIXED, amt = "20", minSub = "200")
        assertEquals(BigDecimal("20.00"), PromotionEngine.discountFor(ok, cart(), now))
        val tooHigh = promo(PromotionType.FIXED, amt = "20", minSub = "999")
        assertNull(PromotionEngine.discountFor(tooHigh, cart(), now))
    }

    @Test
    fun `specific scope only counts its products`() {
        // 10% off product B only -> 10% of 50 = 5.00
        val d = PromotionEngine.discountFor(
            promo(PromotionType.PERCENTAGE, pct = "10", scope = PromotionScope.SPECIFIC, products = listOf(pB)),
            cart(), now,
        )
        assertEquals(BigDecimal("5.00"), d)
    }

    @Test
    fun `bundle needs the bundle size`() {
        // buy 2+ of product A -> RM30 off
        val bundle = promo(PromotionType.BUNDLE, amt = "30", scope = PromotionScope.SPECIFIC, products = listOf(pA), minQty = 2)
        assertEquals(BigDecimal("30.00"), PromotionEngine.discountFor(bundle, cart(), now))
        val bundle3 = promo(PromotionType.BUNDLE, amt = "30", scope = PromotionScope.SPECIFIC, products = listOf(pA), minQty = 3)
        assertNull(PromotionEngine.discountFor(bundle3, cart(), now)) // only 2 of A
    }

    @Test
    fun `expired and not-yet-started promotions do not apply`() {
        val expired = promo(PromotionType.PERCENTAGE, pct = "50", end = now.minus(1, ChronoUnit.HOURS))
        val future = promo(PromotionType.PERCENTAGE, pct = "50", start = now.plus(1, ChronoUnit.HOURS))
        assertNull(PromotionEngine.discountFor(expired, cart(), now))
        assertNull(PromotionEngine.discountFor(future, cart(), now))
    }

    @Test
    fun `best promotion wins - one per order`() {
        val promos = listOf(
            promo(PromotionType.PERCENTAGE, name = "10pct", pct = "10"), // 25.00
            promo(PromotionType.FIXED, name = "RM40", amt = "40", minSub = "200"), // 40.00 (better)
        )
        val best = PromotionEngine.bestPromotion(cart(), promos, now)
        assertEquals("RM40", best?.promotion?.name)
        assertEquals(BigDecimal("40.00"), best?.discount)
    }

    @Test
    fun `discount never exceeds applicable subtotal`() {
        val huge = promo(PromotionType.FIXED, amt = "9999", minSub = "0")
        assertEquals(BigDecimal("250.00"), PromotionEngine.discountFor(huge, cart(), now)) // capped at subtotal
    }
}
