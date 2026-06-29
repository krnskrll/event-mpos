package com.eventmpos.mpos.data.event

import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.auth.toDomain
import com.eventmpos.mpos.data.local.EventSnapshotDao
import com.eventmpos.mpos.data.local.EventSnapshotEntity
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.event.EventRepository
import com.eventmpos.mpos.domain.event.SyncResult
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonPrimitive
import retrofit2.HttpException
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class EventRepositoryImpl @Inject constructor(
    private val api: AuthApi,
    private val dao: EventSnapshotDao,
    private val json: Json,
) : EventRepository {

    override suspend fun getAssignedEvents(): Result<List<AssignedEvent>> =
        withContext(Dispatchers.IO) {
            runCatching { api.getMyEvents().map { it.toDomain() } }
        }

    override suspend fun cachedVersion(eventId: String): String? =
        withContext(Dispatchers.IO) { dao.versionOf(eventId) }

    override suspend fun syncSnapshot(eventId: String): SyncResult =
        withContext(Dispatchers.IO) {
            try {
                val cached = dao.versionOf(eventId)
                val envelope = api.getSnapshot(eventId, cached)

                // Server says nothing changed -> our cached snapshot is already current.
                if (envelope.changed == false) {
                    return@withContext SyncResult.Ready(version = envelope.version, unchanged = true)
                }

                // Full snapshot -> cache it (whole envelope JSON; later epics normalize into tables).
                dao.upsert(
                    EventSnapshotEntity(
                        eventId = eventId,
                        version = envelope.version,
                        snapshotJson = json.encodeToString(
                            com.eventmpos.mpos.data.auth.SnapshotEnvelopeDto.serializer(),
                            envelope,
                        ),
                        syncedAtEpochMs = System.currentTimeMillis(),
                    ),
                )
                SyncResult.Ready(version = envelope.version, unchanged = false)
            } catch (e: HttpException) {
                when (e.code()) {
                    409 -> SyncResult.NotActive(errorMessage(e) ?: "This event is not active yet.")
                    403 -> SyncResult.Failed("You are not assigned to this event.")
                    else -> SyncResult.Failed(errorMessage(e) ?: "Sync failed (${e.code()}).")
                }
            } catch (e: IOException) {
                SyncResult.Failed("Can't reach the server. Check the venue network and retry.")
            }
        }

    /** Pull the human message out of the backend's error JSON body, if present. */
    private fun errorMessage(e: HttpException): String? =
        runCatching {
            val body = e.response()?.errorBody()?.string() ?: return null
            (json.parseToJsonElement(body) as? JsonObject)?.get("message")?.jsonPrimitive?.content
        }.getOrNull()
}
