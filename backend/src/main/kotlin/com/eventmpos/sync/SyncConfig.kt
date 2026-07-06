package com.eventmpos.sync

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table
import java.time.Instant

/** The two schedulable syncs. Stored as the PK so there's exactly one config row per sync. */
enum class SyncType { ERP, MEMBER }

/**
 * Runtime-editable sync schedule, stored in the DB (not application.yml) so an admin can change the
 * cadence or pause a sync without a redeploy. The [DynamicSyncScheduler] reads these rows on every
 * trigger evaluation, so edits take effect on the next tick.
 */
@Entity
@Table(name = "sync_config")
class SyncConfig(
    @Id
    @Column(name = "sync_type", nullable = false)
    val syncType: String,

    @Column(nullable = false)
    var enabled: Boolean,

    @Column(nullable = false)
    var cron: String,

    @Column(name = "updated_at", nullable = false)
    var updatedAt: Instant = Instant.now(),
)
