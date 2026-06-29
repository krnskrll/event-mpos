package com.eventmpos.event

import com.eventmpos.common.ApiException
import com.eventmpos.product.ProductService
import com.eventmpos.user.Role
import com.eventmpos.user.UserRepository
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.security.MessageDigest
import java.time.Instant
import java.util.UUID

@Service
class EventService(
    private val eventRepository: EventRepository,
    private val assignmentRepository: BaEventAssignmentRepository,
    private val userRepository: UserRepository,
    private val productService: ProductService,
) {

    fun listEvents(): List<EventResponse> = eventRepository.findAll().map { it.toResponse() }

    fun getEvent(id: UUID): EventResponse = findOrThrow(id).toResponse()

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
            status = EventStatus.DRAFT,
        )
        return eventRepository.save(event).toResponse()
    }

    @Transactional
    fun updateEvent(id: UUID, request: UpdateEventRequest): EventResponse {
        val event = findOrThrow(id)
        if (event.status == EventStatus.CLOSED) {
            throw ApiException(HttpStatus.CONFLICT, "A closed event can no longer be edited")
        }
        request.name?.let { event.name = it.trim() }
        request.startDate?.let { event.startDate = it }
        request.endDate?.let { event.endDate = it }
        event.updatedAt = Instant.now()
        return eventRepository.save(event).toResponse()
    }

    @Transactional
    fun changeStatus(id: UUID, target: EventStatus): EventResponse {
        val event = findOrThrow(id)
        event.status.assertCanTransitionTo(target)
        event.status = target
        event.updatedAt = Instant.now()
        return eventRepository.save(event).toResponse()
    }

    // ---- BA assignment (from EPIC-01) ----

    @Transactional
    fun assignBaToEvent(userId: UUID, eventId: UUID): EventResponse {
        val user = userRepository.findById(userId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "User not found")
        }
        if (user.role != Role.BA) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Only BA accounts can be assigned to an event")
        }
        val event = findOrThrow(eventId)
        assignmentRepository.findByUserIdAndActiveTrue(userId)?.let { current ->
            if (current.eventId == eventId) return event.toResponse()
            current.active = false
            assignmentRepository.save(current)
        }
        assignmentRepository.save(BaEventAssignment(userId = userId, eventId = eventId))
        return event.toResponse()
    }

    fun activeEventForUser(userId: UUID): EventResponse? =
        assignmentRepository.findByUserIdAndActiveTrue(userId)
            ?.let { eventRepository.findById(it.eventId).orElse(null)?.toResponse() }

    /** All events a BA is assigned to (history + current). Used by the app to decide auto-select. */
    fun assignedEventsForUser(userId: UUID): List<EventResponse> =
        assignmentRepository.findAllByUserId(userId)
            .mapNotNull { eventRepository.findById(it.eventId).orElse(null)?.toResponse() }

    // ---- Snapshot ----

    /**
     * Builds the event snapshot for a BA. Enforces the lifecycle rule: a snapshot (and therefore
     * selling) is only available while the event is ACTIVE.
     */
    fun snapshotForEvent(eventId: UUID): EventSnapshot {
        val event = findOrThrow(eventId)
        requireActive(event)
        return EventSnapshot(
            event = event.toResponse(),
            version = versionOf(event),
            products = productService.availableSnapshotProducts(eventId),
        )
    }

    /** The current snapshot version, used by the device to decide whether a re-sync is needed. */
    fun snapshotVersion(eventId: UUID): String = versionOf(findOrThrow(eventId))

    fun requireAssigned(userId: UUID, eventId: UUID) {
        val assigned = assignmentRepository.findAllByUserId(userId).any { it.eventId == eventId }
        if (!assigned) throw ApiException(HttpStatus.FORBIDDEN, "You are not assigned to this event")
    }

    private fun requireActive(event: Event) {
        if (event.status != EventStatus.ACTIVE) {
            throw ApiException(
                HttpStatus.CONFLICT,
                "Event '${event.code}' is ${event.status} — selling is only allowed while ACTIVE",
            )
        }
    }

    /**
     * Content hash of everything that should trigger a device re-sync. Today that's the event's own
     * mutable fields; as later epics add products/stock/members/promotions, fold their latest
     * updatedAt / row counts into this string and the version changes automatically.
     */
    private fun versionOf(event: Event): String {
        val material = listOf(
            event.id,
            event.name,
            event.startDate,
            event.endDate,
            event.status,
            event.updatedAt.toEpochMilli(),
            // products available for this event — so adding/removing/repricing a product re-syncs devices
            productService.availabilityFingerprint(event.id),
        ).joinToString("|")
        val digest = MessageDigest.getInstance("SHA-256").digest(material.toByteArray())
        return digest.joinToString("") { "%02x".format(it) }.take(16)
    }

    private fun findOrThrow(id: UUID): Event =
        eventRepository.findById(id).orElseThrow { ApiException(HttpStatus.NOT_FOUND, "Event not found") }
}
