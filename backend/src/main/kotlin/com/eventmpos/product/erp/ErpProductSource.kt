package com.eventmpos.product.erp

/**
 * The pluggable ERP boundary. Swapping between an API-pull and a file-import is just a different
 * implementation of this one method — the sync service and everything downstream stay unchanged
 * (per EPIC-03: "easy to swap between API-pull and file-import without rewriting everything").
 *
 * Implementations are selected by the `app.erp.source` property:
 *   mock (default, dev) -> [MockErpProductSource]
 *   file                -> [FileErpProductSource]   (CSV/Excel export)
 *   rest                -> [RestErpProductSource]    (ERP REST API)
 */
interface ErpProductSource {
    /** A human label for logs / the sync status (e.g. "mock", "file:products.csv"). */
    val description: String

    /** Pull the full current product list from the ERP. Throws on transport/parse failure. */
    fun fetchProducts(): List<ErpProduct>
}
