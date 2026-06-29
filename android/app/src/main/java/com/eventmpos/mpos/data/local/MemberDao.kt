package com.eventmpos.mpos.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction

@Dao
interface MemberDao {
    /** O(1) primary-key lookup — the instant offline QR identification path. */
    @Query("SELECT * FROM member WHERE qrCode = :qrCode LIMIT 1")
    suspend fun findByQr(qrCode: String): MemberEntity?

    @Query("SELECT COUNT(*) FROM member")
    suspend fun count(): Int

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(items: List<MemberEntity>)

    @Query("DELETE FROM member")
    suspend fun clear()

    @Transaction
    suspend fun replaceAll(items: List<MemberEntity>) {
        clear()
        insertAll(items)
    }
}
