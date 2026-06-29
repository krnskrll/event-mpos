package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.PrimaryKey

/**
 * The cached event snapshot. We store the version and the raw snapshot JSON; later epics will
 * normalize products/stock/members/promotions into their own Room tables, but the version-gated
 * sync pipeline is established here. Backend remains the source of truth — this is only a cache.
 */
@Entity(tableName = "event_snapshot")
data class EventSnapshotEntity(
    @PrimaryKey val eventId: String,
    val version: String,
    val snapshotJson: String,
    val syncedAtEpochMs: Long,
)
