package com.eventmpos.event

/**
 * The one-call payload the Android app downloads when a BA selects an event (per .cursorrules:
 * "download products + stock + members + promotions to Room in one call").
 *
 * Those collections are populated by later epics (EPIC-03 products/stock, EPIC-04 members,
 * EPIC-06 promotions). For now they're present but empty so the Android sync pipeline + Room
 * caching + version handling can be built and tested end-to-end against a stable shape.
 *
 * [version] is a content hash. The device stores it after a sync; on the next sync it sends the
 * stored version and the server returns 304-style `changed=false` when nothing moved, so the BA
 * only re-downloads when the event data actually changed.
 */
data class EventSnapshot(
    val event: EventResponse,
    val version: String,
    val products: List<Any> = emptyList(),
    val stock: List<Any> = emptyList(),
    val members: List<Any> = emptyList(),
    val promotions: List<Any> = emptyList(),
)

/** Returned by the conditional snapshot endpoint when the client's version is already current. */
data class SnapshotStatus(
    val changed: Boolean,
    val version: String,
)
