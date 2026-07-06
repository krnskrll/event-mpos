package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.PrimaryKey

/**
 * The rendered receipt, stored after an order confirms so it can be reprinted later with NO network
 * call. Keyed by the order's idempotency key (same key as the pending_order row).
 */
@Entity(tableName = "receipt")
data class ReceiptEntity(
    @PrimaryKey val orderKey: String,
    val eventId: String,
    val orderNumberDisplay: String,
    val receiptText: String,        // newline-joined 58mm lines
    val printedOk: Boolean,
    val createdAtEpochMs: Long,
    /** Backend order UUID — used to log print/reprint events to the server audit. */
    val backendOrderId: String? = null,
)
