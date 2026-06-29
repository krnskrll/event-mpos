package com.eventmpos.me

import com.eventmpos.auth.AuthPrincipal
import com.eventmpos.event.EventResponse
import com.eventmpos.event.EventService
import com.eventmpos.user.Role
import com.eventmpos.common.ApiException
import org.springframework.http.HttpStatus
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

data class MeResponse(
    val id: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
)

/**
 * "Who am I" + the BA's assigned event. The Android app calls /api/me/event right after login to
 * know which event to load a snapshot for (EPIC-02+). Available to any authenticated user.
 */
@RestController
@RequestMapping("/api/me")
class MeController(
    private val eventService: EventService,
) {

    @GetMapping
    fun me(@AuthenticationPrincipal principal: AuthPrincipal): MeResponse =
        MeResponse(principal.userId, principal.email, principal.fullName, principal.role)

    /**
     * The active event assigned to the calling BA. Returns 404 if the caller has no active
     * assignment, and 400 if a non-BA calls it (only BAs are event-scoped this way).
     */
    @GetMapping("/event")
    fun myEvent(@AuthenticationPrincipal principal: AuthPrincipal): EventResponse {
        if (principal.role != Role.BA) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Only BA accounts are assigned to a single event")
        }
        return eventService.activeEventForUser(principal.userId)
            ?: throw ApiException(HttpStatus.NOT_FOUND, "No event is assigned to your account yet")
    }
}
