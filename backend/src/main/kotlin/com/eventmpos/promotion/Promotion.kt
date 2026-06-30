package com.eventmpos.promotion

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

enum class PromotionType { PERCENTAGE, FIXED, BUNDLE }

/** Whether the promotion looks at the whole cart or only specific products. */
enum class PromotionScope { ALL, SPECIFIC }

/**
 * An event-scoped promotion that auto-applies to the cart. Promotions are evaluated by the shared
 * [PromotionEngine] (same logic client + server); the server is authoritative at order submit.
 *
 * Phase 1: one promotion per order (engine picks the best), types PERCENTAGE / FIXED / BUNDLE.
 */
@Entity
@Table(name = "promotions", indexes = [Index(name = "idx_promo_event", columnList = "event_id")])
class Promotion(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(nullable = false)
    var name: String,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var type: PromotionType,

    /** For PERCENTAGE: percent off (e.g. 10.0 = 10%). */
    @Column(precision = 5, scale = 2)
    var percentOff: BigDecimal? = null,

    /** For FIXED / BUNDLE: amount off in currency units. */
    @Column(name = "amount_off", precision = 12, scale = 2)
    var amountOff: BigDecimal? = null,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var scope: PromotionScope = PromotionScope.ALL,

    /** Comma-separated product UUIDs for SPECIFIC scope / BUNDLE. */
    @Column(name = "product_ids", length = 4000)
    var productIdsCsv: String? = null,

    /** Minimum total quantity of applicable products (also the bundle size for BUNDLE). */
    @Column(name = "min_qty")
    var minQty: Int? = null,

    /** Minimum applicable subtotal required. */
    @Column(name = "min_subtotal", precision = 12, scale = 2)
    var minSubtotal: BigDecimal? = null,

    /** Optional active window within the event. */
    @Column(name = "starts_at")
    var startsAt: Instant? = null,

    @Column(name = "ends_at")
    var endsAt: Instant? = null,

    @Column(nullable = false)
    var active: Boolean = true,

    /** Tie-breaker when two promotions yield the same discount (higher wins). */
    @Column(nullable = false)
    var priority: Int = 0,

    @Column(name = "updated_at", nullable = false)
    var updatedAt: Instant = Instant.now(),
) {
    var productIds: List<UUID>
        get() = productIdsCsv?.split(",")?.filter { it.isNotBlank() }?.map { UUID.fromString(it.trim()) } ?: emptyList()
        set(value) { productIdsCsv = value.joinToString(",") { it.toString() }.ifBlank { null } }
}
