package com.eventmpos.event

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table
import java.time.Instant
import java.time.LocalDate
import java.util.UUID

/**
 * An Event is the central domain boundary of the whole system (see .cursorrules).
 * In this epic we only need enough of it to assign BAs and show the assigned event after login;
 * later epics (EPIC-02) flesh out event management.
 */
@Entity
@Table(name = "events")
class Event(
    @Id
    val id: UUID = UUID.randomUUID(),

    /** Short human code used in reference ids, e.g. "EV01". Unique. */
    @Column(nullable = false, unique = true)
    var code: String,

    @Column(nullable = false)
    var name: String,

    @Column(name = "start_date")
    var startDate: LocalDate? = null,

    @Column(name = "end_date")
    var endDate: LocalDate? = null,

    @Column(nullable = false)
    var active: Boolean = true,

    @Column(name = "created_at", nullable = false)
    val createdAt: Instant = Instant.now(),
)
