package com.eventmpos.mpos.data.promotion

import com.eventmpos.mpos.data.local.PromotionDao
import com.eventmpos.mpos.data.local.PromotionEntity
import com.eventmpos.mpos.domain.promotion.Promotion
import com.eventmpos.mpos.domain.promotion.PromotionRepository
import com.eventmpos.mpos.domain.promotion.PromotionScope
import com.eventmpos.mpos.domain.promotion.PromotionType
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import java.time.Instant
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PromotionRepositoryImpl @Inject constructor(
    private val dao: PromotionDao,
) : PromotionRepository {

    override fun observePromotions(eventId: String): Flow<List<Promotion>> =
        dao.observeForEvent(eventId).map { list -> list.map { it.toDomain() } }

    private fun PromotionEntity.toDomain() = Promotion(
        id = id,
        name = name,
        type = runCatching { PromotionType.valueOf(type) }.getOrDefault(PromotionType.PERCENTAGE),
        percentOff = percentOff,
        amountOff = amountOff,
        scope = runCatching { PromotionScope.valueOf(scope) }.getOrDefault(PromotionScope.ALL),
        productIds = productIds.split(",").filter { it.isNotBlank() },
        minQty = minQty,
        minSubtotal = minSubtotal,
        startsAtEpochMs = startsAt?.let { runCatching { Instant.parse(it).toEpochMilli() }.getOrNull() },
        endsAtEpochMs = endsAt?.let { runCatching { Instant.parse(it).toEpochMilli() }.getOrNull() },
    )
}
