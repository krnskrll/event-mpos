package com.eventmpos.admin

import com.eventmpos.event.AssignBaRequest
import com.eventmpos.event.ChangeStatusRequest
import com.eventmpos.event.CreateEventRequest
import com.eventmpos.event.EventResponse
import com.eventmpos.event.EventService
import com.eventmpos.event.UpdateEventRequest
import com.eventmpos.user.CreateUserRequest
import com.eventmpos.user.Role
import com.eventmpos.user.UserResponse
import com.eventmpos.user.UserService
import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PatchMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/**
 * Admin-only surface. Every endpoint here lives under the "/api/admin" tree, which is locked to
 * ROLE_ADMIN in SecurityConfig, so individual methods don't need to re-check the role.
 */
@RestController
@RequestMapping("/api/admin")
class AdminController(
    private val userService: UserService,
    private val eventService: EventService,
) {

    // ---- User management ----

    @GetMapping("/users")
    fun listUsers(@RequestParam(required = false) role: Role?): List<UserResponse> =
        if (role != null) userService.listByRole(role) else userService.listUsers()

    @PostMapping("/users")
    @ResponseStatus(HttpStatus.CREATED)
    fun createUser(@Valid @RequestBody request: CreateUserRequest): UserResponse =
        userService.createUser(request)

    // ---- Event management (minimal for EPIC-01; expanded in EPIC-02) ----

    @GetMapping("/events")
    fun listEvents(): List<EventResponse> = eventService.listEvents()

    @PostMapping("/events")
    @ResponseStatus(HttpStatus.CREATED)
    fun createEvent(@Valid @RequestBody request: CreateEventRequest): EventResponse =
        eventService.createEvent(request)

    @PutMapping("/events/{id}")
    fun updateEvent(
        @PathVariable id: UUID,
        @Valid @RequestBody request: UpdateEventRequest,
    ): EventResponse = eventService.updateEvent(id, request)

    /** Drive the lifecycle: DRAFT -> ACTIVE -> CLOSED. Invalid transitions return 409. */
    @PatchMapping("/events/{id}/status")
    fun changeStatus(
        @PathVariable id: UUID,
        @Valid @RequestBody request: ChangeStatusRequest,
    ): EventResponse = eventService.changeStatus(id, request.status)

    // ---- BA → event assignment ----

    @PostMapping("/assignments")
    fun assignBa(@Valid @RequestBody request: AssignBaRequest): EventResponse =
        eventService.assignBaToEvent(request.userId, request.eventId)
}
