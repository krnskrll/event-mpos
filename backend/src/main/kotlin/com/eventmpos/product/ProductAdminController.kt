package com.eventmpos.product

import com.eventmpos.product.erp.ErpSyncResult
import com.eventmpos.product.erp.ErpSyncService
import com.eventmpos.product.erp.ErpSyncStatus
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/**
 * Admin-only product + ERP surface (under /api/admin, locked to ROLE_ADMIN in SecurityConfig).
 * Covers ERP sync, manual product CRUD (admins can add & edit products), and choosing which
 * products are available for each event.
 */
@RestController
@RequestMapping("/api/admin")
class ProductAdminController(
    private val productService: ProductService,
    private val erpSyncService: ErpSyncService,
) {

    // ---- catalogue ----

    @GetMapping("/products")
    fun list(): List<ProductResponse> = productService.list()

    @GetMapping("/products/{id}")
    fun get(@PathVariable id: UUID): ProductResponse = productService.get(id)

    @PostMapping("/products")
    @ResponseStatus(HttpStatus.CREATED)
    fun create(@Valid @RequestBody request: CreateProductRequest): ProductResponse =
        productService.create(request)

    @PutMapping("/products/{id}")
    fun update(@PathVariable id: UUID, @Valid @RequestBody request: UpdateProductRequest): ProductResponse =
        productService.update(id, request)

    // ---- ERP sync ----

    @PostMapping("/erp/sync")
    fun syncNow(): ErpSyncResult = erpSyncService.sync()

    @GetMapping("/erp/status")
    fun syncStatus(): ErpSyncStatus = erpSyncService.status()

    // ---- per-event product availability ----

    @GetMapping("/events/{eventId}/products")
    fun eventProducts(@PathVariable eventId: UUID): List<EventProductResponse> =
        productService.eventProducts(eventId)

    @PutMapping("/events/{eventId}/products/{productId}")
    fun setAvailability(
        @PathVariable eventId: UUID,
        @PathVariable productId: UUID,
        @Valid @RequestBody request: SetAvailabilityRequest,
    ): EventProductResponse = productService.setAvailability(eventId, productId, request.available)
}
