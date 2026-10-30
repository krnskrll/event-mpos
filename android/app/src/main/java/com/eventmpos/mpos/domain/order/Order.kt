package com.eventmpos.mpos.domain.order

import com.eventmpos.mpos.domain.receipt.ReceiptLine
import kotlinx.coroutines.flow.Flow

/** PENDING = confirmed on the device (final number, receipt printed) but not yet uploaded. */
enum class OrderState { PENDING, CONFIRMED, FAILED, CANCELLED }

/** An order as the device knows it. */
data class OrderRecord(
    val idempotencyKey: String,
    val customerName: String,
    val state: OrderState,
    val orderNumberDisplay: String?,
    val referenceId: String?,
    val total: Double?,
    val error: String?,
    val createdAtEpochMs: Long,
    /** Uploaded and not cancelled; the server still decides (a BA may only cancel while NEW). */
    val canCancel: Boolean = false,
)

/** The priced cart being confirmed — what the receipt prints (the device's promotion preview). */
data class Sale(
    val items: List<ReceiptLine>,
    val subtotal: Double,
    val discount: Double,
    val promotionName: String?,
    val total: Double,
)

interface OrderRepository {
    /**
     * Confirm the current cart on the device: assign the next final number from this device's block,
     * print the receipt, and queue the upload. Works offline. Returns the order's idempotency key.
     * Throws [com.eventmpos.mpos.data.order.NoOrderNumberException] when the device has no numbers left.
     */
    suspend fun confirmCurrentCart(eventId: String, sale: Sale): String

    /** Orders for the event (pending, confirmed, failed, cancelled), newest first. */
    fun observeOrders(eventId: String): Flow<List<OrderRecord>>

    /** Re-kick the submission worker (e.g. on reconnect) so PENDING orders drain in FIFO order. */
    fun retryPending()

    /** Cancel an uploaded order. Returns null on success, else a message for the BA. */
    suspend fun cancel(orderKey: String): String?
}
