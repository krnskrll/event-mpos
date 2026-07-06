package com.eventmpos.stock

import jakarta.validation.Valid
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/** Admin: view + set per-event allocations (locked once the event leaves DRAFT). */
@RestController
@RequestMapping("/api/admin")
class StockAdminController(
    private val stockService: StockService,
) {
    @GetMapping("/events/{eventId}/stock")
    fun allocations(@PathVariable eventId: UUID): List<AllocationRow> =
        stockService.allocationsForEvent(eventId)

    @PutMapping("/events/{eventId}/stock/{productId}")
    fun setAllocation(
        @PathVariable eventId: UUID,
        @PathVariable productId: UUID,
        @Valid @RequestBody request: SetAllocationRequest,
    ): AllocationRow = stockService.setAllocation(eventId, productId, request.allocatedQty)
}

/** Warehouse team: pick an event + see near-real-time remaining stock (polled). */
@RestController
@RequestMapping("/api/warehouse")
class WarehouseStockController(
    private val stockService: StockService,
    private val eventService: com.eventmpos.event.EventService,
) {
    @GetMapping("/events")
    fun events(): List<com.eventmpos.event.EventResponse> = eventService.listEvents()

    @GetMapping("/events/{eventId}/stock")
    fun levels(@PathVariable eventId: UUID): List<StockLevelRow> =
        stockService.levelsForEvent(eventId)
}
