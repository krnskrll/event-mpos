package com.eventmpos.product

import com.eventmpos.common.ApiException
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.Instant
import java.util.UUID

@Service
class ProductService(
    private val productRepository: ProductRepository,
    private val barcodeRepository: ProductBarcodeRepository,
    private val eventProductRepository: EventProductRepository,
) {

    fun list(): List<ProductResponse> =
        productRepository.findAll().sortedBy { it.sku }.map { it.toResponse() }

    fun get(id: UUID): ProductResponse = findOrThrow(id).toResponse()

    @Transactional
    fun create(request: CreateProductRequest): ProductResponse {
        if (productRepository.existsBySku(request.sku)) {
            throw ApiException(HttpStatus.CONFLICT, "A product with SKU '${request.sku}' already exists")
        }
        request.barcodes.forEach { assertBarcodeFree(it.barcode) }
        val product = Product(
            sku = request.sku.trim(),
            name = request.name.trim(),
            brand = request.brand?.trim(),
            price = request.price,
            imageUrl = request.imageUrl?.trim(),
            source = ProductSource.MANUAL,
        )
        product.replaceBarcodes(request.barcodes.map { it.barcode.trim() to it.label?.trim() })
        return productRepository.save(product).toResponse()
    }

    @Transactional
    fun update(id: UUID, request: UpdateProductRequest): ProductResponse {
        val product = findOrThrow(id)
        request.name?.let { product.name = it.trim() }
        request.brand?.let { product.brand = it.trim() }
        request.price?.let { product.price = it }
        request.imageUrl?.let { product.imageUrl = it.trim() }
        request.active?.let { product.active = it }
        request.barcodes?.let { incoming ->
            // validate any newly-introduced barcode isn't owned by a different product
            incoming.forEach { bc ->
                val owner = productRepository.findByBarcode(bc.barcode.trim())
                if (owner != null && owner.id != product.id) {
                    throw ApiException(HttpStatus.CONFLICT, "Barcode '${bc.barcode}' belongs to another product")
                }
            }
            product.replaceBarcodes(incoming.map { it.barcode.trim() to it.label?.trim() })
        }
        product.updatedAt = Instant.now()
        return productRepository.save(product).toResponse()
    }

    // ---- per-event availability ----

    /** All catalogue products, each flagged with whether it's available for [eventId]. */
    fun eventProducts(eventId: UUID): List<EventProductResponse> {
        val availableIds = eventProductRepository.findAllByEventId(eventId)
            .filter { it.available }.map { it.productId }.toSet()
        return productRepository.findAll().sortedBy { it.sku }.map {
            EventProductResponse(it.toResponse(), it.id in availableIds)
        }
    }

    @Transactional
    fun setAvailability(eventId: UUID, productId: UUID, available: Boolean): EventProductResponse {
        val product = findOrThrow(productId)
        val link = eventProductRepository.findByEventIdAndProductId(eventId, productId)
            ?: EventProduct(eventId = eventId, productId = productId)
        link.available = available
        eventProductRepository.save(link)
        return EventProductResponse(product.toResponse(), available)
    }

    // ---- snapshot / scan support (used by EventService + MeController) ----

    /** Products available for an event, as compact snapshot DTOs (only catalogue-active ones). */
    fun availableSnapshotProducts(eventId: UUID): List<ProductSnapshotDto> {
        val ids = eventProductRepository.findAllByEventIdAndAvailableTrue(eventId).map { it.productId }
        if (ids.isEmpty()) return emptyList()
        return productRepository.findAllById(ids).filter { it.active }.map { it.toSnapshotDto() }
    }

    /** A stable fingerprint of the event's available products, folded into the snapshot version. */
    fun availabilityFingerprint(eventId: UUID): String {
        val products = availableSnapshotProducts(eventId)
        return products.joinToString(",") { "${it.sku}:${it.price}:${it.barcodes.size}" }
    }

    fun findAvailableByBarcode(eventId: UUID, barcode: String): ProductResponse {
        val product = productRepository.findByBarcode(barcode.trim())
            ?: throw ApiException(HttpStatus.NOT_FOUND, "No product found for barcode '$barcode'")
        val link = eventProductRepository.findByEventIdAndProductId(eventId, product.id)
        if (link == null || !link.available || !product.active) {
            throw ApiException(HttpStatus.NOT_FOUND, "That product isn't available for this event")
        }
        return product.toResponse()
    }

    private fun assertBarcodeFree(barcode: String) {
        if (barcodeRepository.existsByBarcode(barcode.trim())) {
            throw ApiException(HttpStatus.CONFLICT, "Barcode '$barcode' is already in use")
        }
    }

    private fun findOrThrow(id: UUID): Product =
        productRepository.findById(id).orElseThrow { ApiException(HttpStatus.NOT_FOUND, "Product not found") }
}
