package com.eventmpos.sync

import com.eventmpos.common.ApiException
import jakarta.annotation.PostConstruct
import org.springframework.beans.factory.annotation.Value
import org.springframework.http.HttpStatus
import org.springframework.scheduling.support.CronExpression
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.Instant

/**
 * Owns the DB-stored sync schedules. Seeds one row per [SyncType] on startup (disabled by default,
 * using the historical application.yml cron defaults) and validates cron strings on update so a bad
 * expression can never break the scheduler.
 */
@Service
class SyncConfigService(
    private val repository: SyncConfigRepository,
    @Value("\${app.erp.sync-cron:0 0 * * * *}") private val erpDefaultCron: String,
    @Value("\${app.member.sync-cron:0 0 2 * * *}") private val memberDefaultCron: String,
) {
    @PostConstruct
    @Transactional
    fun seedDefaults() {
        if (!repository.existsById(SyncType.ERP.name)) {
            repository.save(SyncConfig(SyncType.ERP.name, enabled = false, cron = erpDefaultCron))
        }
        if (!repository.existsById(SyncType.MEMBER.name)) {
            repository.save(SyncConfig(SyncType.MEMBER.name, enabled = false, cron = memberDefaultCron))
        }
    }

    fun get(type: SyncType): SyncConfig =
        repository.findById(type.name).orElseGet {
            val cron = if (type == SyncType.ERP) erpDefaultCron else memberDefaultCron
            repository.save(SyncConfig(type.name, enabled = false, cron = cron))
        }

    fun all(): List<SyncConfig> = SyncType.entries.map { get(it) }

    @Transactional
    fun update(type: SyncType, enabled: Boolean, cron: String): SyncConfig {
        if (!CronExpression.isValidExpression(cron)) {
            throw ApiException(HttpStatus.BAD_REQUEST, "Invalid cron expression: '$cron'")
        }
        val config = get(type)
        config.enabled = enabled
        config.cron = cron.trim()
        config.updatedAt = Instant.now()
        return repository.save(config)
    }
}
