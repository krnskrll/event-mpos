package com.eventmpos.stock

import com.eventmpos.common.ApiException
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.product.Product
import com.eventmpos.product.ProductRepository
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.util.UUID

/** Low-stock threshold for warnings on the dashboard + device. */
private const val LOW_STOCK_THRESHOLD = 10

@Service
class StockService(
    private val allocationRepository: EventStockAllocationRepository,
    private val productRepository: ProductRepository,
    private val eventRepository: EventRepository,
) {

    // ---- admin allocation ----

    fun allocationsForEvent(eventId: UUID): List<AllocationRow> {
        val byProduct = allocationRepository.findAllByEventId(eventId).associateBy { it.productId }
        return productRepository.findAll().sortedBy { it.sku }.map { p ->
            val a = byProduct[p.id]
            AllocationRow(
                productId = p.id, sku = p.sku, name = p.name, warehouseStock = p.warehouseStock,
                allocatedQty = a?.allocatedQty ?: 0, soldQty = a?.soldQty ?: 0, remaining = a?.remaining ?: 0,
            )
        }
    }

    /**
     * Set how much of a product is allocated to an event. Allowed only while the event is DRAFT —
     * once ACTIVE, allocations are locked (EPIC-05 rule). Can't allocate more than the warehouse has,
     * nor below what's already been sold.
     */
    @Transactional
    fun setAllocation(eventId: UUID, productId: UUID, allocatedQty: Int): AllocationRow {
        val event = eventRepository.findById(eventId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Event not found")
        }
        if (event.status != EventStatus.DRAFT) {
            throw ApiException(HttpStatus.CONFLICT, "Allocations are locked once an event leaves DRAFT (status: ${event.status})")
        }
        val product = productRepository.findById(productId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Product not found")
        }
        if (allocatedQty > product.warehouseStock) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Only ${product.warehouseStock} in warehouse for ${product.sku}")
        }
        val allocation = allocationRepository.findByEventIdAndProductId(eventId, productId)
            ?: EventStockAllocation(eventId = eventId, productId = productId)
        if (allocatedQty < allocation.soldQty) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Cannot allocate below already-sold quantity (${allocation.soldQty})")
        }
        allocation.allocatedQty = allocatedQty
        allocationRepository.save(allocation)
        return toRow(product, allocation)
    }

    /** Seed/bootstrap helper — sets an allocation directly, bypassing the DRAFT lock (used by the seeder). */
    @Transactional
    fun forceAllocate(eventId: UUID, productId: UUID, allocatedQty: Int) {
        val allocation = allocationRepository.findByEventIdAndProductId(eventId, productId)
            ?: EventStockAllocation(eventId = eventId, productId = productId)
        allocation.allocatedQty = allocatedQty
        allocationRepository.save(allocation)
    }

    // ---- views ----

    fun levelsForEvent(eventId: UUID): List<StockLevelRow> {
        val products = productRepository.findAll().associateBy { it.id }
        return allocationRepository.findAllByEventId(eventId)
            .mapNotNull { a ->
                val p = products[a.productId] ?: return@mapNotNull null
                StockLevelRow(p.id, p.sku, p.name, a.allocatedQty, a.soldQty, a.remaining, a.remaining in 1..LOW_STOCK_THRESHOLD)
            }
            .sortedBy { it.sku }
    }

    fun baLevelsForEvent(eventId: UUID): List<BaStockLevel> =
        allocationRepository.findAllByEventId(eventId)
            .map { BaStockLevel(it.productId, it.remaining, it.remaining in 1..LOW_STOCK_THRESHOLD) }

    /** Remaining quantity for a single product (used to enrich the snapshot). */
    fun remaining(eventId: UUID, productId: UUID): Int =
        allocationRepository.findByEventIdAndProductId(eventId, productId)?.remaining ?: 0

    /** Fingerprint of allocation *configuration* (not soldQty) — folded into the snapshot version so
     *  changing allocations re-syncs devices, but ordinary sales (soldQty) do NOT force a full re-sync. */
    fun allocationFingerprint(eventId: UUID): String =
        allocationRepository.findAllByEventId(eventId).sortedBy { it.productId }
            .joinToString(",") { "${it.productId}:${it.allocatedQty}" }

    // ---- the atomic, oversell-proof decrement (called by order submission in EPIC-07) ----

    /**
     * Atomically reserve [qty] of a product's event stock. Runs in a single transaction holding a
     * PESSIMISTIC_WRITE row lock, so simultaneous submissions for the same product serialize and
     * overselling is impossible. Throws 409 if insufficient stock remains.
     *
     * EPIC-07 will call this inside the same transaction that creates the order.
     */
    @Transactional
    fun decrement(eventId: UUID, productId: UUID, qty: Int) {
        require(qty > 0) { "qty must be positive" }
        val allocation = allocationRepository.findForUpdate(eventId, productId)
            ?: throw ApiException(HttpStatus.CONFLICT, "No stock allocated for this product at this event")
        if (allocation.remaining < qty) {
            throw ApiException(
                HttpStatus.CONFLICT,
                "Insufficient stock: ${allocation.remaining} remaining, requested $qty",
            )
        }
        allocation.soldQty += qty
        allocationRepository.save(allocation)
    }

    private fun toRow(p: Product, a: EventStockAllocation) =
        AllocationRow(p.id, p.sku, p.name, p.warehouseStock, a.allocatedQty, a.soldQty, a.remaining)
}
