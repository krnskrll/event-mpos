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
import com.eventmpos.product.ProductResponse
import com.eventmpos.product.ProductService
import com.eventmpos.stock.BaStockLevel
import com.eventmpos.stock.StockService
import com.eventmpos.user.Role
import org.springframework.http.HttpStatus
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

data class MeResponse(
    val id: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
)

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
