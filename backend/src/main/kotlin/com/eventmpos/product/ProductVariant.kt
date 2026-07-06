package com.eventmpos.product

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.JoinColumn
import jakarta.persistence.ManyToOne
import jakarta.persistence.Table
import java.math.BigDecimal
import java.util.UUID

/**
 * A sellable variation of a product — e.g. a lipstick shade or a size. Shares the product's event
 * stock allocation (stock is per product, not per variant); a variant only changes what's shown and,
 * optionally, the price ([priceOverride]). Its own [sku] is unique so it can be identified distinctly.
 */
@Entity
@Table(
    name = "product_variants",
    indexes = [Index(name = "idx_product_variant_sku", columnList = "sku", unique = true)],
)
class ProductVariant(
    @Id
    val id: UUID = UUID.randomUUID(),

    @ManyToOne
    @JoinColumn(name = "product_id", nullable = false)
    var product: Product,

    @Column(nullable = false)
    var name: String,

    @Column(nullable = false, unique = true)
    var sku: String,

    /** If set, overrides the parent product's price for this variant; otherwise the product price applies. */
    @Column(name = "price_override", precision = 12, scale = 2)
    var priceOverride: BigDecimal? = null,

    @Column(nullable = false)
    var active: Boolean = true,
) {
    /** The price actually charged for this variant. */
    fun effectivePrice(): BigDecimal = priceOverride ?: product.price
}
