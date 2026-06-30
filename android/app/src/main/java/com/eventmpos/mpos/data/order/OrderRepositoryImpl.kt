package com.eventmpos.mpos.data.order

import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.local.PendingOrderEntity
import com.eventmpos.mpos.domain.order.OrderRecord
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.order.OrderState
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
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
) : OrderRepository {

    override suspend fun confirmCurrentCart(eventId: String): String {
        val items = cartStore.items.value
        val customer = cartStore.customer.value
            ?: error("A customer must be identified before submitting the order")
        require(items.isNotEmpty()) { "Cart is empty" }

        val key = UUID.randomUUID().toString()
        val lines = items.map { (productId, qty) -> OrderLineDto(productId, qty) }
        dao.upsert(
            PendingOrderEntity(
                idempotencyKey = key,
                eventId = eventId,
                customerName = customer.name,
                customerQr = customer.qrCode,
                customerTier = customer.tier,
                linesJson = json.encodeToString(ListSerializer(OrderLineDto.serializer()), lines),
                status = "PENDING",
                createdAtEpochMs = System.currentTimeMillis(),
            ),
        )
        enqueue()
        cartStore.clear()
        return key
    }

    override fun observeOrders(eventId: String): Flow<List<OrderRecord>> =
        dao.observeForEvent(eventId).map { list ->
            list.map {
                OrderRecord(
                    idempotencyKey = it.idempotencyKey,
                    customerName = it.customerName,
                    state = runCatching { OrderState.valueOf(it.status) }.getOrDefault(OrderState.PENDING),
                    orderNumberDisplay = it.orderNumberDisplay,
                    referenceId = it.referenceId,
                    total = it.total,
                    error = it.errorMessage,
                    createdAtEpochMs = it.createdAtEpochMs,
                )
            }
        }

    private fun enqueue() {
        val request = OneTimeWorkRequestBuilder<SubmitOrderWorker>()
            .setConstraints(Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build())
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 10, TimeUnit.SECONDS)
            .build()
        workManager.enqueueUniqueWork("submit-orders", ExistingWorkPolicy.APPEND_OR_REPLACE, request)
    }
}
