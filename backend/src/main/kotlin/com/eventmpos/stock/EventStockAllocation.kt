package com.eventmpos.stock

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import jakarta.persistence.UniqueConstraint
import jakarta.persistence.Version
import java.util.UUID

/**
 * Stock allocated to an event for a product. Once allocated, the platform owns this count (ERP no
 * longer authoritative). [allocatedQty] is set by an admin; [soldQty] is incremented atomically as
 * orders are confirmed. Remaining = allocatedQty - soldQty, and must never go negative.
 */
@Entity
@Table(
    name = "event_stock_allocations",
    uniqueConstraints = [UniqueConstraint(name = "uk_event_stock", columnNames = ["event_id", "product_id"])],
    indexes = [Index(name = "idx_event_stock_event", columnList = "event_id")],
)
class EventStockAllocation(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(name = "product_id", nullable = false)
    val productId: UUID,

    @Column(name = "allocated_qty", nullable = false)
    var allocatedQty: Int = 0,

    @Column(name = "sold_qty", nullable = false)
    var soldQty: Int = 0,

    /** Optimistic-lock guard (belt-and-braces alongside the pessimistic row lock on decrement). */
    @Version
    @Column(nullable = false)
    var version: Long = 0,
) {
    val remaining: Int get() = allocatedQty - soldQty
}
