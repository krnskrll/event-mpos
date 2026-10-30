package com.eventmpos.mpos.domain.promotion

import kotlinx.coroutines.flow.Flow

enum class PromotionType { PERCENTAGE, FIXED, BUNDLE }
enum class PromotionScope { ALL, SPECIFIC }

/** Cached event promotion (mirrors the backend Promotion fields needed for evaluation/display). */
data class Promotion(
    val id: String,
    val name: String,
    val type: PromotionType,
    val percentOff: Double?,
    val amountOff: Double?,
    val scope: PromotionScope,
    val productIds: List<String>,
    val minQty: Int?,
    val minSubtotal: Double?,
    val startsAtEpochMs: Long?,
    val endsAtEpochMs: Long?,
    /** Tie-breaker between equal discounts (higher wins), as on the server. */
    val priority: Int = 0,
) {
    /** Human description for the promotions screen. */
    fun describe(): String {
        val base = when (type) {
            PromotionType.PERCENTAGE -> "${percentOff?.toInt() ?: 0}% off"
            PromotionType.FIXED -> "RM${amountOff ?: 0.0} off"
            PromotionType.BUNDLE -> "Buy ${minQty ?: 0}, RM${amountOff ?: 0.0} off"
        }
        val cond = buildList {
            minSubtotal?.let { add("min RM$it") }
            if (type != PromotionType.BUNDLE) minQty?.let { add("min qty $it") }
        }.joinToString(", ")
        return if (cond.isBlank()) base else "$base · $cond"
    }
}

interface PromotionRepository {
    /** Active/cached promotions for an event, from Room (offline). */
    fun observePromotions(eventId: String): Flow<List<Promotion>>
}
