package com.eventmpos.audit

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.time.Instant
import java.util.UUID

enum class AuditAction { ORDER_CREATED, STATUS_CHANGED }

/**
 * Append-only audit trail of order activity. Rows are only ever inserted — never updated or deleted
 * — so the history of who did what, and when, is immutable. Every field is a `val` set at
 * construction; there are no mutators. Indexed by event (for the per-event audit view) and by order
 * (for a single order's history).
 */
@Entity
@Table(
    name = "order_audit_log",
    indexes = [
        Index(name = "idx_audit_event", columnList = "event_id"),
        Index(name = "idx_audit_order", columnList = "order_id"),
    ],
)
class OrderAuditLog(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(name = "order_id", nullable = false)
    val orderId: UUID,

    @Column(name = "order_reference", nullable = false)
    val orderReference: String,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    val action: AuditAction,

    @Column(name = "actor_user_id")
    val actorUserId: UUID? = null,

    @Column(name = "actor_name", nullable = false)
    val actorName: String,

    @Column(name = "from_status")
    val fromStatus: String? = null,

    @Column(name = "to_status")
    val toStatus: String? = null,

    @Column(name = "detail")
    val detail: String? = null,

    @Column(name = "created_at", nullable = false)
    val createdAt: Instant = Instant.now(),
)
