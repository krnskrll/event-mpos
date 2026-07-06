package com.eventmpos.mpos.domain.product

import kotlinx.coroutines.flow.Flow

/** A sellable variation (shade/size). Price is the effective (charged) price. */
data class ProductVariant(
    val id: String,
    val name: String,
    val price: Double,
)

/** Domain product for the BA selling flow (price in major currency units). */
data class Product(
    val id: String,
    val sku: String,
    val name: String,
    val brand: String?,
    val price: Double,
    val imageUrl: String?,
    val barcodes: List<String>,
    val remainingStock: Int,
    val variants: List<ProductVariant> = emptyList(),
) {
    val outOfStock: Boolean get() = remainingStock <= 0
    val lowStock: Boolean get() = remainingStock in 1..10
    val hasVariants: Boolean get() = variants.isNotEmpty()
}

interface ProductRepository {
    /** Live list of products for an event, served from Room (offline-friendly). */
    fun observeProducts(eventId: String): Flow<List<Product>>

    /** Resolve a scanned barcode to a product within the event, or null if unknown. */
    suspend fun findByBarcode(eventId: String, barcode: String): Product?

    /** A single product by id (for the detail screen), or null. */
    suspend fun getProduct(eventId: String, productId: String): Product?

    /** Poll the server for live remaining stock and update the Room cache. Best-effort (offline = no-op). */
    suspend fun refreshStock(eventId: String)
}
