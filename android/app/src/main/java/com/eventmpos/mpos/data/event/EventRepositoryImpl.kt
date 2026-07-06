package com.eventmpos.mpos.data.event

import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.auth.toDomain
import com.eventmpos.mpos.data.local.EventSnapshotDao
import com.eventmpos.mpos.data.local.EventSnapshotEntity
import com.eventmpos.mpos.data.local.MemberDao
import com.eventmpos.mpos.data.local.MemberEntity
import com.eventmpos.mpos.data.local.ProductDao
import com.eventmpos.mpos.data.local.ProductEntity
import com.eventmpos.mpos.data.local.PromotionDao
import com.eventmpos.mpos.data.local.PromotionEntity
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
    private val productDao: ProductDao,
    private val memberDao: MemberDao,
    private val promotionDao: PromotionDao,
    private val json: Json,
) : EventRepository {

    override suspend fun getAssignedEvents(): Result<List<AssignedEvent>> =
        withContext(Dispatchers.IO) {
            runCatching { api.getMyEvents().map { it.toDomain() } }
        }

    override suspend fun cachedVersion(eventId: String): String? =
        withContext(Dispatchers.IO) { dao.versionOf(eventId) }

    override suspend fun snapshotHash(eventId: String): String? =
        withContext(Dispatchers.IO) { runCatching { api.getSnapshotHash(eventId).snapshotHash }.getOrNull() }

    override suspend fun syncSnapshot(eventId: String): SyncResult =
        withContext(Dispatchers.IO) {
            try {
                val cached = dao.versionOf(eventId)
                val envelope = api.getSnapshot(eventId, cached)

                // Server says nothing changed -> our cached snapshot is already current.
                if (envelope.changed == false) {
                    return@withContext SyncResult.Ready(version = envelope.version, unchanged = true)
                }

                // Full snapshot -> cache it. Normalize products into their own table for fast
                // offline browsing/scanning; keep the raw envelope for anything not yet normalized.
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
                productDao.replaceForEvent(
                    eventId,
                    envelope.products.map { p ->
                        ProductEntity(
                            rowKey = "$eventId:${p.id}",
                            eventId = eventId,
                            id = p.id,
                            sku = p.sku,
                            name = p.name,
                            brand = p.brand,
                            price = p.price,
                            imageUrl = p.imageUrl,
                            barcodes = p.barcodes.joinToString("\n"),
                            remainingStock = p.remainingStock,
                            variantsJson = json.encodeToString(
                                kotlinx.serialization.builtins.ListSerializer(com.eventmpos.mpos.data.auth.VariantDto.serializer()),
                                p.variants,
                            ),
                        )
                    },
                )
                // Member directory is global; refresh the device cache for offline QR identification.
                memberDao.replaceAll(
                    envelope.members.map { m -> MemberEntity(qrCode = m.qrCode, name = m.name, tier = m.tier) },
                )
                promotionDao.replaceForEvent(
                    eventId,
                    envelope.promotions.map { p ->
                        PromotionEntity(
                            rowKey = "$eventId:${p.id}", eventId = eventId, id = p.id, name = p.name,
                            type = p.type, percentOff = p.percentOff, amountOff = p.amountOff, scope = p.scope,
                            productIds = p.productIds.joinToString(","), minQty = p.minQty,
                            minSubtotal = p.minSubtotal, startsAt = p.startsAt, endsAt = p.endsAt,
                        )
                    },
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
