package com.eventmpos.mpos.domain.order

import kotlinx.coroutines.flow.Flow

enum class OrderState { PENDING, CONFIRMED, FAILED }

/** A submitted order as the device knows it (queued, confirmed, or permanently failed). */
data class OrderRecord(
    val idempotencyKey: String,
    val customerName: String,
    val state: OrderState,
    val orderNumberDisplay: String?,
    val referenceId: String?,
    val total: Double?,
    val error: String?,
    val createdAtEpochMs: Long,
)

interface OrderRepository {
    /** Persist the current cart as a PENDING order and enqueue submission. Returns its idempotency key. */
    suspend fun confirmCurrentCart(eventId: String): String

    /** Submitted orders for the event (pending + confirmed + failed), newest first. */
    fun observeOrders(eventId: String): Flow<List<OrderRecord>>
}
