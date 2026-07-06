package com.eventmpos.order

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.LockModeType
import jakarta.persistence.Table
import jakarta.persistence.UniqueConstraint
import jakarta.persistence.Version
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Lock
import org.springframework.data.jpa.repository.Query
import org.springframework.data.repository.query.Param
import java.util.UUID

/**
 * Per-event, per-day counter for the human order number (#042). Incremented under a pessimistic row
 * lock inside the order transaction so concurrent submissions get strictly increasing, gap-safe
 * numbers (no two orders share #N on the same event+day).
 */
@Entity
@Table(
    name = "order_daily_sequences",
    uniqueConstraints = [UniqueConstraint(name = "uk_seq_event_date", columnNames = ["event_id", "date_key"])],
)
class OrderDailySequence(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(name = "date_key", nullable = false)
    val dateKey: String,

    @Column(name = "last_number", nullable = false)
    var lastNumber: Int = 0,

    @Version
    @Column(nullable = false)
    var version: Long = 0,
)

interface OrderDailySequenceRepository : JpaRepository<OrderDailySequence, UUID> {
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("select s from OrderDailySequence s where s.eventId = :eventId and s.dateKey = :dateKey")
    fun findForUpdate(@Param("eventId") eventId: UUID, @Param("dateKey") dateKey: String): OrderDailySequence?
}
