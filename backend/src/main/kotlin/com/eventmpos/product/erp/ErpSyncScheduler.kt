package com.eventmpos.product.erp

import org.slf4j.LoggerFactory
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.scheduling.annotation.Scheduled
import org.springframework.stereotype.Component

/**
 * Runs the ERP product sync on a configurable schedule (per EPIC-03: "a backend sync job pulls
 * product data from ERP on a configurable schedule").
 *
 * Enable with `app.erp.scheduled-sync-enabled=true` and set the cadence with `app.erp.sync-cron`
 * (default: every hour on the hour). Admins can also trigger a sync on demand via the API.
 */
@Component
@ConditionalOnProperty(name = ["app.erp.scheduled-sync-enabled"], havingValue = "true")
class ErpSyncScheduler(
    private val erpSyncService: ErpSyncService,
) {
    private val log = LoggerFactory.getLogger(ErpSyncScheduler::class.java)

    @Scheduled(cron = "\${app.erp.sync-cron:0 0 * * * *}")
    fun scheduledSync() {
        log.info("Scheduled ERP sync triggered")
        erpSyncService.sync()
    }
}
