package com.eventmpos.mpos.domain.event

import com.eventmpos.mpos.domain.auth.AssignedEvent

/** Outcome of selecting an event and syncing its snapshot to the device. */
sealed interface SyncResult {
    /** Snapshot is on the device and ready. [unchanged] = server reported nothing changed since last sync. */
    data class Ready(val version: String, val unchanged: Boolean) : SyncResult

    /** The event isn't ACTIVE yet, so selling/sync is blocked (lifecycle rule). */
    data class NotActive(val message: String) : SyncResult

    /** Sync failed (network/server). The cached snapshot, if any, is still usable. */
    data class Failed(val message: String) : SyncResult
}

interface EventRepository {
    /** All events assigned to the logged-in BA. */
    suspend fun getAssignedEvents(): Result<List<AssignedEvent>>

    /**
     * Download (or confirm) the snapshot for an event and cache it in Room. Sends the locally
     * cached version so the server can answer "unchanged" and skip a needless re-download.
     */
    suspend fun syncSnapshot(eventId: String): SyncResult

    /** The version of the snapshot currently cached on the device, or null if never synced. */
    suspend fun cachedVersion(eventId: String): String?

    /** The server's current snapshot hash (lightweight freshness check), or null if unreachable. */
    suspend fun snapshotHash(eventId: String): String?
}
