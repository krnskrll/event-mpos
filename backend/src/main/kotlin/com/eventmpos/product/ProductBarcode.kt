package com.eventmpos.product

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.JoinColumn
import jakarta.persistence.ManyToOne
import jakarta.persistence.Table
import java.util.UUID

/**
 * One scannable barcode for a product. A product may have many (e.g. single unit, box, bundle).
 * Barcode values are globally unique so a scan resolves to exactly one product.
 */
@Entity
@Table(name = "product_barcodes")
class ProductBarcode(
    @Id
    val id: UUID = UUID.randomUUID(),

    @ManyToOne
    @JoinColumn(name = "product_id", nullable = false)
    var product: Product,

    @Column(nullable = false, unique = true)
    var barcode: String,

    /** Optional human label for the variant this barcode represents, e.g. "box of 6". */
    var label: String? = null,
)
