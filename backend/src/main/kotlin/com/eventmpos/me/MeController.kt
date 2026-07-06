package com.eventmpos.me

import com.eventmpos.auth.AuthPrincipal
import com.eventmpos.common.ApiException
import com.eventmpos.event.EventResponse
import com.eventmpos.event.EventService
import com.eventmpos.event.EventSnapshot
import com.eventmpos.event.SnapshotStatus
import com.eventmpos.member.MemberResponse
import com.eventmpos.member.MemberService
import com.eventmpos.member.MemberSnapshotDto
import com.eventmpos.order.OrderConfirmation
import com.eventmpos.order.OrderService
import com.eventmpos.order.OrderSummary
import com.eventmpos.order.SubmitOrderRequest
import com.eventmpos.product.ProductResponse
import com.eventmpos.product.ProductService
import com.eventmpos.promotion.PromotionService
import com.eventmpos.receipt.ReceiptLogService
import com.eventmpos.promotion.QuoteRequest
import com.eventmpos.promotion.QuoteResponse
import com.eventmpos.stock.BaStockLevel
import com.eventmpos.stock.StockService
import com.eventmpos.user.Role
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestController
import java.time.Instant
import java.util.UUID

data class MeResponse(
    val id: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
)

/** Lightweight snapshot freshness check — the device compares this hash before a full re-sync. */
data class SnapshotHashResponse(val snapshotHash: String, val checkedAt: Instant)

/**
 * "Who am I" + the BA's event(s) and the event snapshot. The Android app uses these after login to
 * pick/auto-select the assigned event and download its snapshot before any selling.
 */
@RestController
@RequestMapping("/api/me")
class MeController(
    private val eventService: EventService,
    private val productService: ProductService,
    private val memberService: MemberService,
    private val stockService: StockService,
    private val promotionService: PromotionService,
    private val orderService: OrderService,
    private val receiptLogService: ReceiptLogService,
) {

    @GetMapping
    fun me(@AuthenticationPrincipal principal: AuthPrincipal): MeResponse =
        MeResponse(principal.userId, principal.email, principal.fullName, principal.role)

    /** The single active event assigned to the calling BA (EPIC-01). */
    @GetMapping("/event")
    fun myEvent(@AuthenticationPrincipal principal: AuthPrincipal): EventResponse {
        requireBa(principal)
        return eventService.activeEventForUser(principal.userId)
            ?: throw ApiException(HttpStatus.NOT_FOUND, "No event is assigned to your account yet")
    }

    /** All events assigned to the BA. The app auto-selects when there is exactly one. */
    @GetMapping("/events")
    fun myEvents(@AuthenticationPrincipal principal: AuthPrincipal): List<EventResponse> {
        requireBa(principal)
        return eventService.assignedEventsForUser(principal.userId)
    }

    /**
     * Download the full snapshot for an assigned event. Requires the event to be ACTIVE and
     * assigned to the caller. If the client passes its stored [version] and nothing changed, returns
     * a lightweight {changed:false} instead of the whole payload (so devices skip needless re-syncs).
     */
    @GetMapping("/events/{eventId}/snapshot")
    fun snapshot(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
        @RequestParam(required = false) version: String?,
    ): Any {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        if (version != null && version == eventService.snapshotVersion(eventId)) {
            return SnapshotStatus(changed = false, version = version)
        }
        return eventService.snapshotForEvent(eventId)
    }

    /**
     * Just the snapshot hash — the device calls this on pull-to-refresh; if it matches the stored
     * hash, the full snapshot download is skipped entirely.
     */
    @GetMapping("/events/{eventId}/snapshot/hash")
    fun snapshotHash(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
    ): SnapshotHashResponse {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return SnapshotHashResponse(eventService.snapshotVersion(eventId), Instant.now())
    }

    /**
     * Log that a receipt was printed (or reprinted) for an order. Fire-and-forget from Android —
     * printing happens on the device; this is just the backend audit counter. Returns 204.
     */
    @PostMapping("/orders/{orderId}/receipts")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    fun logReceipt(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable orderId: UUID,
    ) {
        requireBa(principal)
        receiptLogService.logPrint(orderId)
    }

    /**
     * Resolve a scanned barcode to a product available for this event. The Android app does this
     * locally from Room while offline; this endpoint backs online lookups and validation.
     */
    @GetMapping("/events/{eventId}/products/by-barcode/{barcode}")
    fun productByBarcode(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
        @PathVariable barcode: String,
    ): ProductResponse {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return productService.findAvailableByBarcode(eventId, barcode)
    }

    /**
     * The active member directory, for the device to cache (the "separate step" alternative to
     * carrying members in the snapshot, useful when the dataset is large).
     */
    @GetMapping("/members")
    fun members(@AuthenticationPrincipal principal: AuthPrincipal): List<MemberSnapshotDto> {
        requireBa(principal)
        return memberService.activeSnapshotMembers()
    }

    /**
     * Live remaining stock per product for the BA's event. The Android app polls this periodically
     * (stock changes faster than the version-gated snapshot, so it's served separately).
     */
    @GetMapping("/events/{eventId}/stock")
    fun stock(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
    ): List<BaStockLevel> {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return stockService.baLevelsForEvent(eventId)
    }

    /**
     * Authoritative cart quote: the server picks + applies the single best promotion (same engine
     * the device previews with). EPIC-07's order submit reuses this so the charged discount is
     * server-decided, not client-supplied.
     */
    @PostMapping("/events/{eventId}/promotions/quote")
    fun quote(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
        @RequestBody request: QuoteRequest,
    ): QuoteResponse {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return promotionService.quote(eventId, request.lines)
    }

    /** Submit a cart → create the order atomically (stock + promo + number). Idempotent per key. */
    @PostMapping("/events/{eventId}/orders")
    fun submitOrder(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
        @Valid @RequestBody request: SubmitOrderRequest,
    ): OrderConfirmation {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return orderService.submit(eventId, principal.userId, request)
    }

    /** This BA's order history for the event session. */
    @GetMapping("/events/{eventId}/orders")
    fun orderHistory(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable eventId: UUID,
    ): List<OrderSummary> {
        requireBa(principal)
        eventService.requireAssigned(principal.userId, eventId)
        return orderService.history(eventId, principal.userId)
    }

    /** Full order detail / re-print payload. */
    @GetMapping("/orders/{orderId}")
    fun order(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable orderId: UUID,
    ): OrderConfirmation {
        requireBa(principal)
        return orderService.get(orderId)
    }

    /** Online QR fallback when a scanned member isn't in the device cache. 404 if unknown/inactive. */
    @GetMapping("/members/by-qr/{qrCode}")
    fun memberByQr(
        @AuthenticationPrincipal principal: AuthPrincipal,
        @PathVariable qrCode: String,
    ): MemberResponse {
        requireBa(principal)
        return memberService.byQr(qrCode)
    }

    private fun requireBa(principal: AuthPrincipal) {
        if (principal.role != Role.BA) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Only BA accounts are scoped to an event")
        }
    }
}
