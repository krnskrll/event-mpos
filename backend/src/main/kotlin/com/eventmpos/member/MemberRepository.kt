package com.eventmpos.member

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface MemberRepository : JpaRepository<Member, UUID> {
    fun findByQrCode(qrCode: String): Member?
    fun findByMemberCode(memberCode: String): Member?
    fun findAllByActiveTrue(): List<Member>
    fun countByActiveTrue(): Long
}
