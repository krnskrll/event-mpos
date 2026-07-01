package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow

@Dao
interface ProductDao {
    @Query("SELECT * FROM product WHERE eventId = :eventId ORDER BY name")
    fun observeForEvent(eventId: String): Flow<List<ProductEntity>>

    @Query("SELECT * FROM product WHERE eventId = :eventId AND id = :productId LIMIT 1")
    suspend fun find(eventId: String, productId: String): ProductEntity?

    @Query("SELECT * FROM product WHERE eventId = :eventId")
    suspend fun forEvent(eventId: String): List<ProductEntity>

    /** Match a scanned barcode within an event (barcodes stored newline-joined). */
    @Query(
        "SELECT * FROM product WHERE eventId = :eventId AND (" +
            "barcodes = :code OR barcodes LIKE :code || char(10) || '%' " +
            "OR barcodes LIKE '%' || char(10) || :code || char(10) || '%' " +
            "OR barcodes LIKE '%' || char(10) || :code) LIMIT 1",
    )
    suspend fun findByBarcode(eventId: String, code: String): ProductEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(items: List<ProductEntity>)

    /** Update just the live remaining stock for a product (from the periodic /stock poll). */
    @Query("UPDATE product SET remainingStock = :remaining WHERE eventId = :eventId AND id = :productId")
    suspend fun updateRemaining(eventId: String, productId: String, remaining: Int)

    @Query("DELETE FROM product WHERE eventId = :eventId")
    suspend fun clearEvent(eventId: String)

    /** Replace the whole product set for an event in one transaction (snapshot sync). */
    @Transaction
    suspend fun replaceForEvent(eventId: String, items: List<ProductEntity>) {
        clearEvent(eventId)
        insertAll(items)
    }
}
