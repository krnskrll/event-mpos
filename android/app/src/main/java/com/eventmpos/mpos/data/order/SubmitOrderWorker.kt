package com.eventmpos.mpos.data.order

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.receipt.ReceiptManager
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import kotlinx.serialization.json.Json
import retrofit2.HttpException
import java.io.IOException

/**
 * Drains PENDING orders to the server. Retried by WorkManager (network constraint + backoff) until
 * each is CONFIRMED. Submitting with the stored idempotency key means retries never duplicate.
 *  - success            -> mark CONFIRMED (store order number / reference / total)
 *  - network/5xx        -> Result.retry()  (BA keeps seeing "Pending")
 *  - 4xx (e.g. sold out)-> mark FAILED with the server message (permanent; needs BA action)
 */
@HiltWorker
class SubmitOrderWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted params: WorkerParameters,
    private val api: AuthApi,
    private val dao: PendingOrderDao,
    private val receiptManager: ReceiptManager,
    private val json: Json,
) : CoroutineWorker(appContext, params) {

    override suspend fun doWork(): Result {
        val pending = dao.pending()
        var shouldRetry = false
        for (order in pending) {
            try {
                val lines = json.decodeFromString(
                    kotlinx.serialization.builtins.ListSerializer(OrderLineDto.serializer()),
                    order.linesJson,
                )
                val resp = api.submitOrder(
                    order.eventId,
                    SubmitOrderRequestDto(
                        idempotencyKey = order.idempotencyKey,
                        customer = CustomerInfoDto(order.customerName, order.customerQr, order.customerTier),
                        lines = lines,
                    ),
                )
                dao.upsert(
                    order.copy(
                        status = "CONFIRMED",
                        orderNumberDisplay = resp.orderNumberDisplay,
                        referenceId = resp.referenceId,
                        total = resp.total,
                        errorMessage = null,
                    ),
                )
                // Auto-print the receipt the moment the server confirms (never on a submit attempt).
                // A printer failure is non-fatal: the receipt is stored for reprint; order stands.
                runCatching { receiptManager.storeAndPrint(order.idempotencyKey, order.eventId, resp) }
            } catch (e: IOException) {
                shouldRetry = true // offline / timeout — keep it pending and retry later
            } catch (e: HttpException) {
                if (e.code() in 500..599) {
                    shouldRetry = true
                } else {
                    dao.upsert(order.copy(status = "FAILED", errorMessage = "Rejected (${e.code()})"))
                }
            }
        }
        return if (shouldRetry) Result.retry() else Result.success()
    }
}
