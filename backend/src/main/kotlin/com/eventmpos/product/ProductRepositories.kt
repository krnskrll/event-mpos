package com.eventmpos.product

import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import org.springframework.data.repository.query.Param
import java.util.UUID

interface ProductRepository : JpaRepository<Product, UUID> {
    fun findBySku(sku: String): Product?
    fun existsBySku(sku: String): Boolean

    @Query("select p from Product p join p.barcodes b where b.barcode = :barcode")
    fun findByBarcode(@Param("barcode") barcode: String): Product?
}

interface ProductBarcodeRepository : JpaRepository<ProductBarcode, UUID> {
    fun existsByBarcode(barcode: String): Boolean
}

interface EventProductRepository : JpaRepository<EventProduct, UUID> {
    fun findByEventIdAndProductId(eventId: UUID, productId: UUID): EventProduct?
    fun findAllByEventId(eventId: UUID): List<EventProduct>
    fun findAllByEventIdAndAvailableTrue(eventId: UUID): List<EventProduct>
}
