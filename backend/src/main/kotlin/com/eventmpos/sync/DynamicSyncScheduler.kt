package com.eventmpos.sync

import com.eventmpos.member.sync.MemberSyncService
import com.eventmpos.product.erp.ErpSyncService
import org.slf4j.LoggerFactory
import org.springframework.scheduling.Trigger
import org.springframework.scheduling.TriggerContext
import org.springframework.scheduling.annotation.SchedulingConfigurer
import org.springframework.scheduling.config.ScheduledTaskRegistrar
import org.springframework.scheduling.support.CronExpression
import org.springframework.stereotype.Component
import java.time.Instant
import java.time.ZoneId
import java.time.ZonedDateTime
import java.util.concurrent.Executors

/**
 * Drives the ERP + member syncs from the DB-stored [SyncConfig] instead of a fixed @Scheduled cron,
 * so admins can change the cadence or pause a sync at runtime. On each trigger evaluation it re-reads
 * the config: the next fire time is always computed from the current cron (edits apply on the next
 * tick), and the task itself is a no-op while the sync is disabled — so toggling enabled on/off never
 * requires a restart and never permanently stops the trigger.
 */
@Component
class DynamicSyncScheduler(
    private val syncConfigService: SyncConfigService,
    private val erpSyncService: ErpSyncService,
    private val memberSyncService: MemberSyncService,
) : SchedulingConfigurer {

    private val log = LoggerFactory.getLogger(DynamicSyncScheduler::class.java)

    override fun configureTasks(registrar: ScheduledTaskRegistrar) {
        registrar.setScheduler(Executors.newScheduledThreadPool(2))
        registrar.addTriggerTask({ runIfEnabled(SyncType.ERP) { erpSyncService.sync() } }, trigger(SyncType.ERP))
        registrar.addTriggerTask({ runIfEnabled(SyncType.MEMBER) { memberSyncService.sync() } }, trigger(SyncType.MEMBER))
    }

    private fun runIfEnabled(type: SyncType, block: () -> Unit) {
        if (syncConfigService.get(type).enabled) {
            log.info("Scheduled {} sync triggered", type)
            runCatching { block() }.onFailure { log.warn("Scheduled {} sync failed: {}", type, it.message) }
        }
    }

    private fun trigger(type: SyncType) = Trigger { ctx: TriggerContext ->
        val cron = syncConfigService.get(type).cron
        val base = ctx.lastCompletion() ?: Instant.now()
        runCatching {
            CronExpression.parse(cron)
                .next(ZonedDateTime.ofInstant(base, ZoneId.systemDefault()))
                ?.toInstant()
        }.getOrNull()
    }
}
