package com.eventmpos.audit

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface OrderAuditLogRepository : JpaRepository<OrderAuditLog, UUID> {
    fun findAllByEventIdOrderByCreatedAtDesc(eventId: UUID): List<OrderAuditLog>
    fun findAllByOrderIdOrderByCreatedAtAsc(orderId: UUID): List<OrderAuditLog>
}
