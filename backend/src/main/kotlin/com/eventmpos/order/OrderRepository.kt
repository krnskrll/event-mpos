package com.eventmpos.order

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface OrderRepository : JpaRepository<Order, UUID> {
    fun findByIdempotencyKey(idempotencyKey: String): Order?
    fun findAllByEventIdAndBaUserIdOrderByCreatedAtDesc(eventId: UUID, baUserId: UUID): List<Order>
    fun findAllByEventIdOrderByCreatedAtDesc(eventId: UUID): List<Order>
}
