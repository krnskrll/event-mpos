package com.eventmpos.mpos.data.order

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import androidx.room.Dao
import androidx.room.Entity
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Transaction
import com.eventmpos.mpos.data.auth.AuthApi
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

/** A range of order numbers the server reserved for this device (EPIC-12 C1, decision D1). */
@Entity(tableName = "number_block", primaryKeys = ["eventId", "rangeStart"])
data class NumberBlockEntity(
    val eventId: String,
    val businessDate: String, // yyyyMMdd the numbers belong to (goes into the order reference)
    val rangeStart: Int,
    val rangeEnd: Int,
    val nextUnused: Int,
)

@Dao
abstract class NumberBlockDao {
    @Query("SELECT * FROM number_block WHERE eventId = :eventId AND nextUnused <= rangeEnd ORDER BY rangeStart LIMIT 1")
    abstract suspend fun current(eventId: String): NumberBlockEntity?

    @Query("SELECT COALESCE(SUM(rangeEnd - nextUnused + 1), 0) FROM number_block WHERE eventId = :eventId AND nextUnused <= rangeEnd")
    abstract suspend fun remaining(eventId: String): Int

    @Insert
    abstract suspend fun insert(block: NumberBlockEntity)

    @Query("UPDATE number_block SET nextUnused = nextUnused + 1 WHERE eventId = :eventId AND rangeStart = :rangeStart")
    abstract suspend fun advance(eventId: String, rangeStart: Int)

    /** Atomically hands out the next unused number (the returned row's nextUnused), or null if none left. */
    @Transaction
    open suspend fun take(eventId: String): NumberBlockEntity? = current(eventId)?.also { advance(eventId, it.rangeStart) }
}

private val Context.devicePrefs by preferencesDataStore("device")

/** A random id generated once per install. A reinstall is a new device, so blocks are never reused. */
@Singleton
open class DeviceIdentity @Inject constructor(@ApplicationContext private val context: Context) {
    open suspend fun id(): String {
        val key = stringPreferencesKey("device_id")
        context.devicePrefs.edit { if (it[key] == null) it[key] = UUID.randomUUID().toString() }
        return context.devicePrefs.data.first()[key]!!
    }
}

class NoOrderNumberException : IllegalStateException(
    "This device has run out of order numbers. Connect to the internet to get more, then confirm again.",
)

/** Final order numbers for offline confirmation, drawn from server-issued per-device blocks. */
@Singleton
class OrderNumbers @Inject constructor(
    private val dao: NumberBlockDao,
    private val api: AuthApi,
    private val device: DeviceIdentity,
) {
    private val topUpLock = Mutex()

    /** The next final number and its business date, or [NoOrderNumberException] when all blocks are used. */
    suspend fun next(eventId: String): Pair<Int, String> =
        dao.take(eventId)?.let { it.nextUnused to it.businessDate } ?: throw NoOrderNumberException()

    /** Fetch a fresh block when fewer than [LOW_WATER] numbers remain. Best effort: offline is fine. */
    suspend fun topUp(eventId: String) = topUpLock.withLock {
        if (dao.remaining(eventId) >= LOW_WATER) return@withLock
        runCatching { api.issueNumberBlock(eventId, NumberBlockRequestDto(device.id())) }
            .onSuccess { dao.insert(NumberBlockEntity(eventId, it.businessDate, it.rangeStart, it.rangeEnd, it.rangeStart)) }
    }

    suspend fun deviceId() = device.id()

    private companion object {
        const val LOW_WATER = 10
    }
}
