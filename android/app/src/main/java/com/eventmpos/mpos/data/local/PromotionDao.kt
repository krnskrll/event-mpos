package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow

@Dao
interface PromotionDao {
    @Query("SELECT * FROM promotion WHERE eventId = :eventId")
    fun observeForEvent(eventId: String): Flow<List<PromotionEntity>>

    @Query("SELECT * FROM promotion WHERE eventId = :eventId")
    suspend fun forEvent(eventId: String): List<PromotionEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(items: List<PromotionEntity>)

    @Query("DELETE FROM promotion WHERE eventId = :eventId")
    suspend fun clearEvent(eventId: String)

    @Transaction
    suspend fun replaceForEvent(eventId: String, items: List<PromotionEntity>) {
        clearEvent(eventId)
        insertAll(items)
    }
}
