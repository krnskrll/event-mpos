package com.eventmpos.member.sync

import com.eventmpos.member.Member
import com.eventmpos.member.MemberRepository
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.Instant

data class MemberSyncResult(
    val ok: Boolean,
    val created: Int,
    val updated: Int,
    val total: Int,
    val source: String,
    val message: String? = null,
)

/**
 * Pulls members from the active [MemberSource] and upserts them by memberCode. Idempotent.
 * An optional region filter (`app.member.region-filter`) lets large datasets be narrowed at sync
 * time — only members from that region are stored.
 */
@Service
class MemberSyncService(
    private val source: MemberSource,
    private val memberRepository: MemberRepository,
    private val statusRepository: MemberSyncStatusRepository,
    @Value("\${app.member.region-filter:}") private val regionFilter: String,
) {
    private val log = LoggerFactory.getLogger(MemberSyncService::class.java)

    @Transactional
    fun sync(): MemberSyncResult {
        log.info("Member sync starting from source: {}", source.description)
        return try {
            val records = source.fetchMembers()
                .filter { regionFilter.isBlank() || it.region.equals(regionFilter, ignoreCase = true) }
            var created = 0
            var updated = 0
            records.forEach { rec ->
                val existing = memberRepository.findByMemberCode(rec.memberCode)
                if (existing == null) {
                    memberRepository.save(
                        Member(
                            memberCode = rec.memberCode,
                            qrCode = rec.qrCode,
                            name = rec.name,
                            tier = rec.tier,
                            region = rec.region,
                            phone = rec.phone,
                            active = rec.active,
                        ),
                    )
                    created++
                } else {
                    existing.qrCode = rec.qrCode
                    existing.name = rec.name
                    existing.tier = rec.tier
                    existing.region = rec.region
                    existing.phone = rec.phone
                    existing.active = rec.active
                    existing.updatedAt = Instant.now()
                    memberRepository.save(existing)
                    updated++
                }
            }
            val result = MemberSyncResult(true, created, updated, records.size, source.description)
            saveStatus(result)
            log.info("Member sync done: {} created, {} updated, {} total", created, updated, records.size)
            result
        } catch (ex: Exception) {
            log.error("Member sync failed", ex)
            val result = MemberSyncResult(false, 0, 0, 0, source.description, ex.message ?: "sync failed")
            saveStatus(result)
            result
        }
    }

    fun status(): MemberSyncStatus = statusRepository.findById(1).orElse(MemberSyncStatus())

    private fun saveStatus(result: MemberSyncResult) {
        val status = statusRepository.findById(1).orElse(MemberSyncStatus())
        status.lastSyncAt = Instant.now()
        status.lastResult = if (result.ok) "SUCCESS" else "FAILED"
        status.memberCount = result.total
        status.source = result.source
        status.message = result.message ?: "created=${result.created}, updated=${result.updated}"
        statusRepository.save(status)
    }
}
