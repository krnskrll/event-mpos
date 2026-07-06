package com.eventmpos.order

import jakarta.persistence.CascadeType
import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.FetchType
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.OneToMany
import jakarta.persistence.Table
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

/** Order lifecycle (warehouse drives NEW→…→COLLECTED in EPIC-09). Orders start NEW. */
enum class OrderStatus {
    NEW, PACKING, READY, COLLECTED,
    ;

    /** The next status forward, or null if already COLLECTED. */
    fun next(): OrderStatus? = entries.getOrNull(ordinal + 1)

    /** Warehouse advances strictly forward, one step at a time. */
    fun canAdvanceTo(target: OrderStatus): Boolean = target == next()
}

/**
 * A confirmed order. Created only on server submit, atomically with stock decrement + promotion
 * application. Carries three identifiers per .cursorrules: internal [id] (UUID), human [referenceId]
 * (ORD-YYYYMMDD-EVxx-NNN), and the daily [orderNumber] (#042) shown on the receipt.
 *
 * [idempotencyKey] is unique — a retried/double-tapped submit with the same key returns this order
 * instead of creating a duplicate.
 */
@Entity
@Table(
    name = "orders",
    indexes = [
        Index(name = "idx_order_event", columnList = "event_id"),
        Index(name = "idx_order_idem", columnList = "idempotency_key", unique = true),
    ],
)
class Order(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(name = "ba_user_id", nullable = false)
    val baUserId: UUID,

    @Column(name = "idempotency_key", nullable = false, unique = true)
    val idempotencyKey: String,

    @Column(name = "order_number", nullable = false)
    val orderNumber: Int,

    @Column(name = "reference_id", nullable = false)
    val referenceId: String,

    // customer (EPIC-04): an order must be linked to a customer
    @Column(name = "customer_name", nullable = false)
    val customerName: String,

    @Column(name = "customer_qr")
    val customerQr: String? = null,

    @Column(name = "customer_tier")
    val customerTier: String? = null,

    @Column(nullable = false, precision = 12, scale = 2)
    val subtotal: BigDecimal,

    @Column(nullable = false, precision = 12, scale = 2)
    val discount: BigDecimal,

    @Column(nullable = false, precision = 12, scale = 2)
    val total: BigDecimal,

    @Column(name = "promotion_id")
    val promotionId: UUID? = null,

    @Column(name = "promotion_name")
    val promotionName: String? = null,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var status: OrderStatus = OrderStatus.NEW,

    @OneToMany(mappedBy = "order", cascade = [CascadeType.ALL], orphanRemoval = true, fetch = FetchType.EAGER)
    val items: MutableList<OrderItem> = mutableListOf(),

    @Column(name = "created_at", nullable = false)
    val createdAt: Instant = Instant.now(),
)
