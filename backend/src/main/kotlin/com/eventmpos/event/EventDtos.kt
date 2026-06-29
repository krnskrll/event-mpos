package com.eventmpos.event

import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import java.time.LocalDate
import java.util.UUID

data class EventResponse(
    val id: UUID,
    val code: String,
    val name: String,
    val startDate: LocalDate?,
    val endDate: LocalDate?,
    val status: EventStatus,
)

data class CreateEventRequest(
    @field:NotBlank val code: String,
    @field:NotBlank val name: String,
    val startDate: LocalDate? = null,
    val endDate: LocalDate? = null,
)

/** Partial edit of mutable event fields. Status changes go through the dedicated status endpoint. */
data class UpdateEventRequest(
    val name: String? = null,
    val startDate: LocalDate? = null,
    val endDate: LocalDate? = null,
)

data class ChangeStatusRequest(
    @field:NotNull val status: EventStatus,
)

data class AssignBaRequest(
    @field:NotNull val userId: UUID,
    @field:NotNull val eventId: UUID,
)

fun Event.toResponse() = EventResponse(id, code, name, startDate, endDate, status)
