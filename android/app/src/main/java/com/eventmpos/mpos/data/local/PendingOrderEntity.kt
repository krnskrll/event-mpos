package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.PrimaryKey

/**
 * An order the BA confirmed on-device. Persisted immediately (so it survives app death / offline)
 * and submitted by [com.eventmpos.mpos.data.order.SubmitOrderWorker] via WorkManager. status:
 * PENDING (queued/sending) → CONFIRMED (server accepted, number assigned) | FAILED (permanent).
 * [idempotencyKey] is the primary key, so retries never create duplicates server-side.
 */
@Entity(tableName = "pending_order")
data class PendingOrderEntity(
    @PrimaryKey val idempotencyKey: String,
    val eventId: String,
    val customerName: String,
    val customerQr: String?,
    val customerTier: String?,
    val linesJson: String,          // serialized List<OrderLineDto>
    val status: String,             // PENDING | CONFIRMED | FAILED
    val createdAtEpochMs: Long,
    // filled once confirmed:
    val orderNumberDisplay: String? = null,
    val referenceId: String? = null,
    val total: Double? = null,
    val errorMessage: String? = null,
    val orderNumber: Int? = null,      // final number from the device's block (EPIC-12 C1)
    val serverOrderId: String? = null, // set once uploaded; needed to cancel
)
