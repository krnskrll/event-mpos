package com.eventmpos.event

import jakarta.validation.constraints.NotBlank
import java.time.LocalDate
import java.util.UUID

data class EventResponse(
    val id: UUID,
    val code: String,
    val name: String,
    val startDate: LocalDate?,
    val endDate: LocalDate?,
    val active: Boolean,
)

data class CreateEventRequest(
    @field:NotBlank val code: String,
    @field:NotBlank val name: String,
    val startDate: LocalDate? = null,
    val endDate: LocalDate? = null,
)

data class AssignBaRequest(
    @field:jakarta.validation.constraints.NotNull val userId: UUID,
    @field:jakarta.validation.constraints.NotNull val eventId: UUID,
)

fun Event.toResponse() = EventResponse(id, code, name, startDate, endDate, active)
