package com.eventmpos.mpos.data.order

import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.local.PendingOrderEntity
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.map

/** In-memory [PendingOrderDao] with the same ordering as the Room queries. */
class FakePendingOrderDao : PendingOrderDao {
    val rows = MutableStateFlow<Map<String, PendingOrderEntity>>(emptyMap())

    override fun observeForEvent(eventId: String): Flow<List<PendingOrderEntity>> =
        rows.map { all -> all.values.filter { it.eventId == eventId }.sortedByDescending { it.createdAtEpochMs } }

    override suspend fun pending(): List<PendingOrderEntity> =
        rows.value.values.filter { it.status == "PENDING" }.sortedBy { it.createdAtEpochMs }

    override suspend fun upsert(order: PendingOrderEntity) {
        rows.value = rows.value + (order.idempotencyKey to order)
    }

    override suspend fun get(key: String): PendingOrderEntity? = rows.value[key]
}
