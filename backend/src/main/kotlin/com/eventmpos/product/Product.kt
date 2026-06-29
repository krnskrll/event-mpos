package com.eventmpos.product

import jakarta.persistence.CascadeType
import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.FetchType
import jakarta.persistence.Id
import jakarta.persistence.OneToMany
import jakarta.persistence.Table
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

/** Where a product came from — pulled from the ERP, or hand-entered by an admin. */
enum class ProductSource { ERP, MANUAL }

/**
 * A catalogue product. Sourced from the ERP (keyed by [sku]) or created manually by an admin.
 * A product can carry several barcodes (variants / packaging) — see [ProductBarcode].
 * Per-event availability lives in [EventProduct], not here, since the catalogue is global.
 */
@Entity
@Table(name = "products")
class Product(
    @Id
    val id: UUID = UUID.randomUUID(),

    /** Stable ERP key (or admin-chosen for manual products). Unique; used as the sync upsert key. */
    @Column(nullable = false, unique = true)
    var sku: String,

    @Column(nullable = false)
    var name: String,

    var brand: String? = null,

    @Column(nullable = false, precision = 12, scale = 2)
    var price: BigDecimal = BigDecimal.ZERO,

    @Column(name = "image_url")
    var imageUrl: String? = null,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var source: ProductSource = ProductSource.MANUAL,

    /** Catalogue-level enable flag. Per-event availability is separate (EventProduct). */
    @Column(nullable = false)
    var active: Boolean = true,

    @OneToMany(mappedBy = "product", cascade = [CascadeType.ALL], orphanRemoval = true, fetch = FetchType.EAGER)
    var barcodes: MutableList<ProductBarcode> = mutableListOf(),

    @Column(name = "created_at", nullable = false)
    val createdAt: Instant = Instant.now(),

    @Column(name = "updated_at", nullable = false)
    var updatedAt: Instant = Instant.now(),
) {
    fun replaceBarcodes(codes: List<Pair<String, String?>>) {
        barcodes.clear()
        codes.forEach { (code, label) ->
            barcodes.add(ProductBarcode(product = this, barcode = code, label = label))
        }
    }
}
