package com.eventmpos.audit

import com.eventmpos.user.UserRepository
import org.springframework.stereotype.Service
import java.time.Instant
import java.util.UUID

/**
 * Records order activity into the append-only [OrderAuditLog]. Resolves the acting user's name once
 * at write time (so the trail stays readable even if the user is later renamed/deleted). Purely
 * additive — it never updates or deletes existing rows.
 */
@Service
class AuditService(
    private val auditRepository: OrderAuditLogRepository,
    private val userRepository: UserRepository,
) {
    private fun actorName(userId: UUID?): String =
        userId?.let { userRepository.findById(it).map { u -> u.fullName }.orElse(null) } ?: "system"

    fun recordOrderCreated(eventId: UUID, orderId: UUID, orderReference: String, actorUserId: UUID?) {
        auditRepository.save(
            OrderAuditLog(
                eventId = eventId, orderId = orderId, orderReference = orderReference,
                action = AuditAction.ORDER_CREATED, actorUserId = actorUserId, actorName = actorName(actorUserId),
                toStatus = "NEW", detail = "Order created",
            ),
        )
    }

    fun recordStatusChange(
        eventId: UUID,
        orderId: UUID,
        orderReference: String,
        from: String,
        to: String,
        actorUserId: UUID?,
    ) {
        auditRepository.save(
            OrderAuditLog(
                eventId = eventId, orderId = orderId, orderReference = orderReference,
                action = AuditAction.STATUS_CHANGED, actorUserId = actorUserId, actorName = actorName(actorUserId),
                fromStatus = from, toStatus = to, detail = "Status $from → $to",
            ),
        )
    }

    fun forEvent(eventId: UUID): List<AuditEntryResponse> =
        auditRepository.findAllByEventIdOrderByCreatedAtDesc(eventId).map { it.toResponse() }

    fun forOrder(orderId: UUID): List<AuditEntryResponse> =
        auditRepository.findAllByOrderIdOrderByCreatedAtAsc(orderId).map { it.toResponse() }
}

data class AuditEntryResponse(
    val id: UUID,
    val orderId: UUID,
    val orderReference: String,
    val action: AuditAction,
    val actorName: String,
    val fromStatus: String?,
    val toStatus: String?,
    val detail: String?,
    val at: Instant,
)

fun OrderAuditLog.toResponse() = AuditEntryResponse(
    id, orderId, orderReference, action, actorName, fromStatus, toStatus, detail, createdAt,
)
