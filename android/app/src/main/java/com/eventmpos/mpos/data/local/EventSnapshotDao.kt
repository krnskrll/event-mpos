package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query

@Dao
interface EventSnapshotDao {
    @Query("SELECT * FROM event_snapshot WHERE eventId = :eventId")
    suspend fun get(eventId: String): EventSnapshotEntity?

    @Query("SELECT version FROM event_snapshot WHERE eventId = :eventId")
    suspend fun versionOf(eventId: String): String?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(entity: EventSnapshotEntity)
}
