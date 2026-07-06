package com.eventmpos.product.erp

import java.math.BigDecimal

/**
 * Normalized product record as it comes out of the ERP, regardless of transport (REST or file).
 * Each [ErpProductSource] is responsible for mapping its raw format into this shape, so the rest of
 * the platform never depends on the ERP's wire format.
 */
data class ErpProduct(
    val sku: String,
    val name: String,
    val brand: String?,
    val price: BigDecimal,
    val imageUrl: String?,
    /** Quantity on hand in the warehouse per the ERP. */
    val stockQuantity: Int,
    /** barcode value -> optional variant label */
    val barcodes: List<Pair<String, String?>>,
)
