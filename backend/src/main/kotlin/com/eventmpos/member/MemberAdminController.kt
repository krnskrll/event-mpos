package com.eventmpos.member

import com.eventmpos.member.sync.MemberSyncResult
import com.eventmpos.member.sync.MemberSyncService
import com.eventmpos.member.sync.MemberSyncStatus
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

/** Admin-only member directory + sync (under /api/admin, ROLE_ADMIN). */
@RestController
@RequestMapping("/api/admin")
class MemberAdminController(
    private val memberService: MemberService,
    private val memberSyncService: MemberSyncService,
) {
    @GetMapping("/members")
    fun list(): List<MemberResponse> = memberService.list()

    @PostMapping("/members/sync")
    fun syncNow(): MemberSyncResult = memberSyncService.sync()

    @GetMapping("/members/sync/status")
    fun status(): MemberSyncStatus = memberSyncService.status()
}
