package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import kotlinx.coroutines.flow.Flow

@Dao
interface PendingOrderDao {
    @Query("SELECT * FROM pending_order WHERE eventId = :eventId ORDER BY createdAtEpochMs DESC")
    fun observeForEvent(eventId: String): Flow<List<PendingOrderEntity>>

    /** Oldest first, so the retry queue uploads in the order the BA confirmed them. */
    @Query("SELECT * FROM pending_order WHERE status = 'PENDING' ORDER BY createdAtEpochMs ASC")
    suspend fun pending(): List<PendingOrderEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(order: PendingOrderEntity)

    @Query("SELECT COUNT(*) FROM pending_order WHERE eventId = :eventId AND status = 'PENDING'")
    fun observePendingCount(eventId: String): Flow<Int>

    @Query("SELECT * FROM pending_order WHERE idempotencyKey = :key")
    suspend fun get(key: String): PendingOrderEntity?
}
