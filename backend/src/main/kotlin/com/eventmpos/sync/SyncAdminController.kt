package com.eventmpos.sync

import com.eventmpos.member.sync.MemberSyncService
import com.eventmpos.product.erp.ErpSyncService
import jakarta.validation.constraints.NotBlank
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import java.time.Instant

data class UpdateSyncConfigRequest(
    val enabled: Boolean,
    @field:NotBlank val cron: String,
)

/** Combined schedule + last-run status for one sync type. */
data class SyncTypeView(
    val type: SyncType,
    val enabled: Boolean,
    val cron: String,
    val updatedAt: Instant,
    val lastSyncAt: Instant?,
    val lastResult: String?,
    val count: Int,
    val source: String?,
    val message: String?,
)

data class SyncOverview(val erp: SyncTypeView, val member: SyncTypeView)

/**
 * Admin-only sync administration (under /api/admin, locked to ROLE_ADMIN). View + edit the ERP and
 * member sync schedules (stored in the DB, changeable at runtime) and see the last run's status.
 * Manual on-demand triggers live at /api/admin/erp/sync and /api/admin/members/sync (EPIC-03/04).
 */
@RestController
@RequestMapping("/api/admin/sync")
class SyncAdminController(
    private val syncConfigService: SyncConfigService,
    private val erpSyncService: ErpSyncService,
    private val memberSyncService: MemberSyncService,
) {
    @GetMapping
    fun overview(): SyncOverview {
        val erpCfg = syncConfigService.get(SyncType.ERP)
        val memCfg = syncConfigService.get(SyncType.MEMBER)
        val erpStatus = erpSyncService.status()
        val memStatus = memberSyncService.status()
        return SyncOverview(
            erp = SyncTypeView(
                SyncType.ERP, erpCfg.enabled, erpCfg.cron, erpCfg.updatedAt,
                erpStatus.lastSyncAt, erpStatus.lastResult, erpStatus.productCount, erpStatus.source, erpStatus.message,
            ),
            member = SyncTypeView(
                SyncType.MEMBER, memCfg.enabled, memCfg.cron, memCfg.updatedAt,
                memStatus.lastSyncAt, memStatus.lastResult, memStatus.memberCount, memStatus.source, memStatus.message,
            ),
        )
    }

    @PutMapping("/{type}")
    fun update(@PathVariable type: SyncType, @RequestBody request: UpdateSyncConfigRequest): SyncTypeView {
        val cfg = syncConfigService.update(type, request.enabled, request.cron)
        val status = if (type == SyncType.ERP) {
            val s = erpSyncService.status(); SyncTypeView(type, cfg.enabled, cfg.cron, cfg.updatedAt, s.lastSyncAt, s.lastResult, s.productCount, s.source, s.message)
        } else {
            val s = memberSyncService.status(); SyncTypeView(type, cfg.enabled, cfg.cron, cfg.updatedAt, s.lastSyncAt, s.lastResult, s.memberCount, s.source, s.message)
        }
        return status
    }
}
