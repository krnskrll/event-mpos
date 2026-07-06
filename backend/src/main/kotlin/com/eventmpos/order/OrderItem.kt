package com.eventmpos.order

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.JoinColumn
import jakarta.persistence.ManyToOne
import jakarta.persistence.Table
import java.math.BigDecimal
import java.util.UUID

/** A line on an order — snapshots name/price at purchase time (catalogue may change later). */
@Entity
@Table(
    name = "order_items",
    // FK columns aren't auto-indexed on Postgres; the end-of-event report joins items by order.
    indexes = [Index(name = "idx_order_item_order", columnList = "order_id")],
)
class OrderItem(
    @Id
    val id: UUID = UUID.randomUUID(),

    @ManyToOne
    @JoinColumn(name = "order_id", nullable = false)
    var order: Order,

    @Column(name = "product_id", nullable = false)
    val productId: UUID,

    /** Chosen variant (shade/size), if any — snapshotted at purchase time. */
    @Column(name = "variant_id")
    val variantId: UUID? = null,

    @Column(name = "variant_name")
    val variantName: String? = null,

    @Column(nullable = false)
    val sku: String,

    @Column(nullable = false)
    val name: String,

    @Column(name = "unit_price", nullable = false, precision = 12, scale = 2)
    val unitPrice: BigDecimal,

    @Column(nullable = false)
    val qty: Int,

    @Column(name = "line_total", nullable = false, precision = 12, scale = 2)
    val lineTotal: BigDecimal,
)
