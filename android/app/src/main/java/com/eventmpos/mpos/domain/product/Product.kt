package com.eventmpos.mpos.domain.product

import kotlinx.coroutines.flow.Flow

/** Domain product for the BA selling flow (price in major currency units). */
data class Product(
    val id: String,
    val sku: String,
    val name: String,
    val brand: String?,
    val price: Double,
    val imageUrl: String?,
    val barcodes: List<String>,
)

interface ProductRepository {
    /** Live list of products for an event, served from Room (offline-friendly). */
    fun observeProducts(eventId: String): Flow<List<Product>>

    /** Resolve a scanned barcode to a product within the event, or null if unknown. */
    suspend fun findByBarcode(eventId: String, barcode: String): Product?
}
