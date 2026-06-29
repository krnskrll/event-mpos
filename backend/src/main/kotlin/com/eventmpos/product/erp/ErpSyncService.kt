package com.eventmpos.product.erp

import com.eventmpos.product.Product
import com.eventmpos.product.ProductRepository
import com.eventmpos.product.ProductSource
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.Instant

data class ErpSyncResult(
    val ok: Boolean,
    val created: Int,
    val updated: Int,
    val total: Int,
    val source: String,
    val message: String? = null,
)

/**
 * Pulls products from whichever [ErpProductSource] is active and upserts them by SKU. Records the
 * outcome in [ErpSyncStatus]. Safe to run repeatedly (idempotent): existing products are updated in
 * place, and barcodes are only rewritten when they actually changed (avoids unique-constraint churn).
 */
@Service
class ErpSyncService(
    private val source: ErpProductSource,
    private val productRepository: ProductRepository,
    private val statusRepository: ErpSyncStatusRepository,
) {
    private val log = LoggerFactory.getLogger(ErpSyncService::class.java)

    @Transactional
    fun sync(): ErpSyncResult {
        log.info("ERP sync starting from source: {}", source.description)
        return try {
            val products = source.fetchProducts()
            var created = 0
            var updated = 0
            products.forEach { erp ->
                val existing = productRepository.findBySku(erp.sku)
                if (existing == null) {
                    val p = Product(
                        sku = erp.sku,
                        name = erp.name,
                        brand = erp.brand,
                        price = erp.price,
                        imageUrl = erp.imageUrl,
                        source = ProductSource.ERP,
                    )
                    p.replaceBarcodes(erp.barcodes)
                    productRepository.save(p)
                    created++
                } else {
                    existing.name = erp.name
                    existing.brand = erp.brand
                    existing.price = erp.price
                    existing.imageUrl = erp.imageUrl
                    existing.source = ProductSource.ERP
                    existing.updatedAt = Instant.now()
                    val incoming = erp.barcodes.map { it.first }.toSet()
                    val current = existing.barcodes.map { it.barcode }.toSet()
                    if (incoming != current) existing.replaceBarcodes(erp.barcodes)
                    productRepository.save(existing)
                    updated++
                }
            }
            val result = ErpSyncResult(true, created, updated, products.size, source.description)
            saveStatus(result)
            log.info("ERP sync done: {} created, {} updated, {} total", created, updated, products.size)
            result
        } catch (ex: Exception) {
            log.error("ERP sync failed", ex)
            val result = ErpSyncResult(false, 0, 0, 0, source.description, ex.message ?: "sync failed")
            saveStatus(result)
            result
        }
    }

    fun status(): ErpSyncStatus = statusRepository.findById(1).orElse(ErpSyncStatus())

    private fun saveStatus(result: ErpSyncResult) {
        val status = statusRepository.findById(1).orElse(ErpSyncStatus())
        status.lastSyncAt = Instant.now()
        status.lastResult = if (result.ok) "SUCCESS" else "FAILED"
        status.productCount = result.total
        status.source = result.source
        status.message = result.message
            ?: "created=${result.created}, updated=${result.updated}, total=${result.total}"
        statusRepository.save(status)
    }
}
