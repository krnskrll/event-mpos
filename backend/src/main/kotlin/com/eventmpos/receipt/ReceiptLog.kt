package com.eventmpos.receipt

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.time.Instant
import java.util.UUID

/**
 * Audit record that a receipt was printed for an order. Printing itself happens on the Android
 * device (Sunmi); this is the backend's minimal role — one row per order, with a reprint counter.
 */
@Entity
@Table(
    name = "receipt_logs",
    indexes = [Index(name = "idx_receipt_log_order", columnList = "order_id", unique = true)],
)
class ReceiptLog(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "order_id", nullable = false, unique = true)
    val orderId: UUID,

    @Column(name = "print_count", nullable = false)
    var printCount: Int = 1,

    @Column(name = "first_printed_at", nullable = false)
    val firstPrintedAt: Instant = Instant.now(),

    @Column(name = "last_printed_at", nullable = false)
    var lastPrintedAt: Instant = Instant.now(),
)
