package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Entity
import androidx.room.Index
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.PrimaryKey
import androidx.room.Query
import kotlinx.coroutines.flow.Flow

/**
 * A cart parked on the device. `live = true` marks the one autosaved copy of the event's current
 * cart (so it survives process death); every other row is a draft the BA saved by name.
 */
@Entity(tableName = "cart_draft", indices = [Index("eventId")])
data class CartDraftEntity(
    @PrimaryKey val id: String,
    val eventId: String,
    val label: String,
    val customerJson: String?,
    val linesJson: String,          // {"<CartKey>": qty}
    val createdAtEpochMs: Long,
    val live: Boolean = false,
)

@Dao
interface CartDraftDao {
    @Query("SELECT * FROM cart_draft WHERE eventId = :eventId AND live = 0 ORDER BY createdAtEpochMs DESC")
    fun observeDrafts(eventId: String): Flow<List<CartDraftEntity>>

    @Query("SELECT * FROM cart_draft WHERE id = :id")
    suspend fun get(id: String): CartDraftEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(draft: CartDraftEntity)

    @Query("DELETE FROM cart_draft WHERE id = :id")
    suspend fun delete(id: String)

    @Query("DELETE FROM cart_draft WHERE eventId = :eventId")
    suspend fun deleteForEvent(eventId: String)
}
