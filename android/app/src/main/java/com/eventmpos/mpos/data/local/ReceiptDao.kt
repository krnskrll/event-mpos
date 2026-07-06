package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query

@Dao
interface ReceiptDao {
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(receipt: ReceiptEntity)

    @Query("SELECT * FROM receipt WHERE orderKey = :orderKey")
    suspend fun get(orderKey: String): ReceiptEntity?

    @Query("UPDATE receipt SET printedOk = :ok WHERE orderKey = :orderKey")
    suspend fun setPrinted(orderKey: String, ok: Boolean)
}
