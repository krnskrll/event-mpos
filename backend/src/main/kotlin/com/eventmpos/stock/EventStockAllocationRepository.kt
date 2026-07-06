package com.eventmpos.stock

import jakarta.persistence.LockModeType
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Lock
import org.springframework.data.jpa.repository.Query
import org.springframework.data.repository.query.Param
import java.util.UUID

interface EventStockAllocationRepository : JpaRepository<EventStockAllocation, UUID> {

    fun findByEventIdAndProductId(eventId: UUID, productId: UUID): EventStockAllocation?

    fun findAllByEventId(eventId: UUID): List<EventStockAllocation>

    /**
     * Pessimistic write lock (PostgreSQL `SELECT ... FOR UPDATE`) on the allocation row. This is the
     * heart of oversell prevention: the decrementing transaction holds the row lock so concurrent
     * order submissions for the same product serialize instead of racing on remaining stock.
     */
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("select a from EventStockAllocation a where a.eventId = :eventId and a.productId = :productId")
    fun findForUpdate(@Param("eventId") eventId: UUID, @Param("productId") productId: UUID): EventStockAllocation?
}
