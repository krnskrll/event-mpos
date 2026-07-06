package com.eventmpos.member.sync

import org.slf4j.LoggerFactory
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.scheduling.annotation.Scheduled
import org.springframework.stereotype.Component

/**
 * Scheduled member pull. Enable with `app.member.scheduled-sync-enabled=true`; cadence via
 * `app.member.sync-cron` (default daily at 02:00). Admins can also sync on demand.
 */
@Component
@ConditionalOnProperty(name = ["app.member.scheduled-sync-enabled"], havingValue = "true")
class MemberSyncScheduler(
    private val memberSyncService: MemberSyncService,
) {
    private val log = LoggerFactory.getLogger(MemberSyncScheduler::class.java)

    @Scheduled(cron = "\${app.member.sync-cron:0 0 2 * * *}")
    fun scheduledSync() {
        log.info("Scheduled member sync triggered")
        memberSyncService.sync()
    }
}
