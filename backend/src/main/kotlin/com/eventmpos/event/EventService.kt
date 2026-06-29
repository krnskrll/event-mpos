package com.eventmpos.event

import com.eventmpos.common.ApiException
import com.eventmpos.user.Role
import com.eventmpos.user.UserRepository
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.util.UUID

@Service
class EventService(
    private val eventRepository: EventRepository,
    private val assignmentRepository: BaEventAssignmentRepository,
    private val userRepository: UserRepository,
) {

    fun listEvents(): List<EventResponse> = eventRepository.findAll().map { it.toResponse() }

    fun getEvent(id: UUID): EventResponse =
        eventRepository.findById(id).orElseThrow { ApiException(HttpStatus.NOT_FOUND, "Event not found") }
            .toResponse()

    @Transactional
    fun createEvent(request: CreateEventRequest): EventResponse {
        if (eventRepository.existsByCode(request.code)) {
            throw ApiException(HttpStatus.CONFLICT, "An event with code '${request.code}' already exists")
        }
        val event = Event(
            code = request.code.trim(),
            name = request.name.trim(),
            startDate = request.startDate,
            endDate = request.endDate,
        )
        return eventRepository.save(event).toResponse()
    }

    /**
     * Assigns a BA to an event. Enforces that the target user is actually a BA and deactivates any
     * prior active assignment so the BA always has exactly one active event (what the app shows).
     */
    @Transactional
    fun assignBaToEvent(userId: UUID, eventId: UUID): EventResponse {
        val user = userRepository.findById(userId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "User not found")
        }
        if (user.role != Role.BA) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Only BA accounts can be assigned to an event")
        }
        val event = eventRepository.findById(eventId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Event not found")
        }
        assignmentRepository.findByUserIdAndActiveTrue(userId)?.let { current ->
            if (current.eventId == eventId) return event.toResponse() // already assigned, no-op
            current.active = false
            assignmentRepository.save(current)
        }
        assignmentRepository.save(BaEventAssignment(userId = userId, eventId = eventId))
        return event.toResponse()
    }

    /** The active event for a BA, or null if none assigned yet. Used by the Android app post-login. */
    fun activeEventForUser(userId: UUID): EventResponse? =
        assignmentRepository.findByUserIdAndActiveTrue(userId)
            ?.let { eventRepository.findById(it.eventId).orElse(null)?.toResponse() }
}
