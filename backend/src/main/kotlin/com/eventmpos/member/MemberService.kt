package com.eventmpos.member

import com.eventmpos.common.ApiException
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service

@Service
class MemberService(
    private val memberRepository: MemberRepository,
) {

    fun list(): List<MemberResponse> =
        memberRepository.findAll().sortedBy { it.memberCode }.map { it.toResponse() }

    fun activeCount(): Long = memberRepository.countByActiveTrue()

    /** Active members for the device cache / snapshot (compact). */
    fun activeSnapshotMembers(): List<MemberSnapshotDto> =
        memberRepository.findAllByActiveTrue().map { it.toSnapshotDto() }

    /** Exact QR lookup — the online fallback when a member isn't in the device cache. */
    fun byQr(qrCode: String): MemberResponse {
        val member = memberRepository.findByQrCode(qrCode.trim())
        if (member == null || !member.active) {
            throw ApiException(HttpStatus.NOT_FOUND, "No active member for QR '$qrCode'")
        }
        return member.toResponse()
    }

    /**
     * Folded into the event snapshot version so member changes trigger a device re-sync.
     * Content-based (not updatedAt) so an idempotent re-sync that changes nothing keeps the same
     * version — devices only re-download when member data actually changed.
     */
    fun fingerprint(): String {
        val members = memberRepository.findAllByActiveTrue().sortedBy { it.qrCode }
        val material = members.joinToString(",") { "${it.qrCode}:${it.name}:${it.tier}" }
        return "${members.size}:${material.hashCode()}"
    }
}
