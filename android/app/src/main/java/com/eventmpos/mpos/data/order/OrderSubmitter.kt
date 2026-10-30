package com.eventmpos.mpos.data.order

import com.eventmpos.mpos.core.notification.OrderNotifier
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.receipt.ReceiptManager
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import retrofit2.HttpException
import java.io.IOException
import java.time.Instant
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Uploads PENDING orders, oldest first. They were already confirmed on the device (final number,
 * receipt printed); upload only syncs them. The stored idempotency key makes retries duplicate-free.
 *  - success             -> CONFIRMED (server id stored, receipt linked for reprint logging)
 *  - network error / 5xx -> stays PENDING; [drainPending] returns true so the worker retries
 *  - other 4xx (sold out)-> FAILED with the reason (permanent; needs BA action)
 */
@Singleton
class OrderSubmitter @Inject constructor(
    private val api: AuthApi,
    private val dao: PendingOrderDao,
    private val receiptManager: ReceiptManager,
    private val notifier: OrderNotifier,
    private val json: Json,
    private val numbers: OrderNumbers,
) {
    /** Returns true when at least one order must be retried later. */
    suspend fun drainPending(): Boolean {
        var shouldRetry = false
        for (order in dao.pending()) {
            try {
                val lines = json.decodeFromString(ListSerializer(OrderLineDto.serializer()), order.linesJson)
                val resp = api.submitOrder(
                    order.eventId,
                    SubmitOrderRequestDto(
                        idempotencyKey = order.idempotencyKey,
                        customer = CustomerInfoDto(order.customerName, order.customerQr, order.customerTier),
                        lines = lines,
                        deviceId = order.orderNumber?.let { numbers.deviceId() },
                        orderNumber = order.orderNumber,
                        clientCreatedAt = Instant.ofEpochMilli(order.createdAtEpochMs).toString(),
                    ),
                )
                dao.upsert(
                    order.copy(
                        status = "CONFIRMED",
                        orderNumberDisplay = resp.orderNumberDisplay,
                        referenceId = resp.referenceId,
                        total = resp.total,
                        errorMessage = null,
                        serverOrderId = resp.id,
                    ),
                )
                runCatching { receiptManager.linkServerOrder(order.idempotencyKey, resp.id) }
                // Let the BA know a queued sale went through, even if they're on another screen.
                notifier.notifyConfirmed(order.idempotencyKey, resp.orderNumberDisplay, resp.total)
            } catch (e: IOException) {
                shouldRetry = true // offline / timeout — keep it pending and retry later
            } catch (e: HttpException) {
                if (e.code() in 500..599) {
                    shouldRetry = true
                } else {
                    val reason = "Rejected (${e.code()})"
                    dao.upsert(order.copy(status = "FAILED", errorMessage = reason))
                    notifier.notifyFailed(order.idempotencyKey, order.customerName, reason)
                }
            }
        }
        return shouldRetry
    }
}
