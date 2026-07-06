package com.eventmpos.order

import com.eventmpos.auth.AuthPrincipal
import jakarta.validation.Valid
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PatchMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/**
 * Warehouse order board API (ROLE_WAREHOUSE/ADMIN — the /api/warehouse tree is locked in SecurityConfig).
 * Real-time updates arrive over STOMP (topic /topic/events/{id}/orders); these REST endpoints back the
 * initial load + the polling fallback + status actions.
 */
@RestController
@RequestMapping("/api/warehouse")
class WarehouseOrderController(
    private val orderService: OrderService,
) {
    @GetMapping("/events/{eventId}/orders")
    fun orders(
        @PathVariable eventId: UUID,
        @RequestParam(required = false) status: OrderStatus?,
    ): List<WarehouseOrderView> = orderService.warehouseOrders(eventId, status)

    @PatchMapping("/orders/{orderId}/status")
    fun advance(
        @PathVariable orderId: UUID,
        @Valid @RequestBody request: AdvanceStatusRequest,
        @AuthenticationPrincipal principal: AuthPrincipal,
    ): WarehouseOrderView = orderService.advanceStatus(orderId, request.status, principal.userId)
}
