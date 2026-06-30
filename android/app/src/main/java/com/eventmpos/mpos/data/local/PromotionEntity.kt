package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey

/** Event promotion cached for offline auto-apply preview. productIds stored comma-joined. */
@Entity(tableName = "promotion", indices = [Index("eventId")])
data class PromotionEntity(
    @PrimaryKey val rowKey: String, // "<eventId>:<id>"
    val eventId: String,
    val id: String,
    val name: String,
    val type: String,
    val percentOff: Double?,
    val amountOff: Double?,
    val scope: String,
    val productIds: String,
    val minQty: Int?,
    val minSubtotal: Double?,
    val startsAt: String?,
    val endsAt: String?,
)
