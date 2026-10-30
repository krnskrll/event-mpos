package com.eventmpos.mpos.data.order

import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import com.eventmpos.mpos.BuildConfig
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.cart.CartKey
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.local.PendingOrderEntity
import com.eventmpos.mpos.data.local.ProductDao
import com.eventmpos.mpos.data.receipt.ReceiptManager
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.order.OrderState
import com.eventmpos.mpos.domain.order.Sale
import com.eventmpos.mpos.domain.receipt.ReceiptData
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import retrofit2.HttpException
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.UUID
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class OrderRepositoryImpl @Inject constructor(
    private val cartStore: CartStore,
    private val dao: PendingOrderDao,
    private val workManager: WorkManager,
    private val json: Json,
    private val numbers: OrderNumbers,
    private val receipts: ReceiptManager,
    private val products: ProductDao,
    private val session: SessionStore,
    private val api: AuthApi,
    private val scope: CoroutineScope,
) : OrderRepository {

    override suspend fun confirmCurrentCart(eventId: String, sale: Sale): String {
        val items = cartStore.items.value
        val customer = cartStore.customer.value
            ?: error("A customer must be identified before submitting the order")
        require(items.isNotEmpty()) { "Cart is empty" }

        // Final number from this device's block (throws when none are left), so the receipt is final now.
        val (number, businessDate) = numbers.next(eventId)
        val display = "#%03d".format(number)
        val reference = "ORD-$businessDate-${session.event.value?.code.orEmpty()}-%03d".format(number)
        val key = UUID.randomUUID().toString()
        val now = System.currentTimeMillis()
        val lines = items.map { (cartKey, qty) -> OrderLineDto(CartKey.productId(cartKey), qty, CartKey.variantId(cartKey)) }
        dao.upsert(
            PendingOrderEntity(
                idempotencyKey = key, eventId = eventId, customerName = customer.name, customerQr = customer.qrCode,
                customerTier = customer.tier, linesJson = json.encodeToString(ListSerializer(OrderLineDto.serializer()), lines),
                status = "PENDING", createdAtEpochMs = now, orderNumber = number, orderNumberDisplay = display,
                referenceId = reference, total = sale.total,
            ),
        )
        lines.forEach { products.adjustRemaining(eventId, it.productId, -it.qty) } // local best guess; server decides
        cartStore.clear()
        enqueue()

        // Print right away. A printer problem never affects the order: the receipt is stored for reprint.
        runCatching {
            receipts.storeAndPrint(
                key, eventId,
                ReceiptData(
                    brand = BuildConfig.BRAND_NAME, eventName = session.event.value?.name.orEmpty(),
                    orderNumberDisplay = display, referenceId = reference, customerName = customer.name,
                    customerTier = customer.tier, items = sale.items, subtotal = sale.subtotal, discount = sale.discount,
                    promotionName = sale.promotionName, total = sale.total, cashierNote = CASHIER_NOTE,
                    dateText = SimpleDateFormat("dd/MM HH:mm", Locale.getDefault()).format(Date(now)),
                ),
            )
        }
        scope.launch { numbers.topUp(eventId) } // keep the block topped up while online
        return key
    }

    override fun observeOrders(eventId: String): Flow<List<OrderRecord>> =
        dao.observeForEvent(eventId).map { list ->
            list.map {
                val state = runCatching { OrderState.valueOf(it.status) }.getOrDefault(OrderState.PENDING)
                OrderRecord(
                    idempotencyKey = it.idempotencyKey, customerName = it.customerName, state = state,
                    orderNumberDisplay = it.orderNumberDisplay, referenceId = it.referenceId, total = it.total,
                    error = it.errorMessage, createdAtEpochMs = it.createdAtEpochMs,
                    canCancel = state == OrderState.CONFIRMED && it.serverOrderId != null,
                )
            }
        }

    override fun retryPending() = enqueue()

    override suspend fun cancel(orderKey: String): String? {
        val order = dao.get(orderKey) ?: return "Order not found"
        val serverId = order.serverOrderId ?: return "This order hasn't synced yet — try again when it's confirmed."
        return try {
            api.cancelOrder(serverId, CancelOrderRequestDto())
            dao.upsert(order.copy(status = OrderState.CANCELLED.name))
            receipts.markCancelled(orderKey)
            json.decodeFromString(ListSerializer(OrderLineDto.serializer()), order.linesJson)
                .forEach { products.adjustRemaining(order.eventId, it.productId, it.qty) }
            null
        } catch (e: HttpException) {
            if (e.code() == 409) "The warehouse is already packing this order — it can't be cancelled here." else "Cancel failed (${e.code()})"
        } catch (e: java.io.IOException) {
            "Can't reach the server. Cancel needs a connection."
        }
    }

    private fun enqueue() {
        val request = OneTimeWorkRequestBuilder<SubmitOrderWorker>()
            .setConstraints(Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build())
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 10, TimeUnit.SECONDS)
            .build()
        workManager.enqueueUniqueWork("submit-orders", ExistingWorkPolicy.APPEND_OR_REPLACE, request)
    }

    private companion object {
        const val CASHIER_NOTE = "Please proceed to the cashier to complete payment."
    }
}
