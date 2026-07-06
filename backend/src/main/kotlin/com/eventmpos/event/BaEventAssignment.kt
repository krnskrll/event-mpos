package com.eventmpos.event

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.time.Instant
import java.util.UUID

/**
 * Maps a BA (user) to the event they are working. A BA can have history across events,
 * but at most one assignment is [active] at a time — the active one is what the Android app
 * loads after login. Admin manages these from the web dashboard.
 *
 * Foreign keys are stored as raw UUIDs (not JPA relations) to keep this entity lightweight and
 * avoid eager-loading the full user/event graphs on the hot login path.
 */
@Entity
@Table(
    name = "ba_event_assignments",
    indexes = [
        Index(name = "idx_assignment_user", columnList = "user_id"),
        Index(name = "idx_assignment_event", columnList = "event_id"),
    ],
)
class BaEventAssignment(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "user_id", nullable = false)
    val userId: UUID,

    @Column(name = "event_id", nullable = false)
    var eventId: UUID,

    @Column(nullable = false)
    var active: Boolean = true,

    @Column(name = "assigned_at", nullable = false)
    val assignedAt: Instant = Instant.now(),
)
