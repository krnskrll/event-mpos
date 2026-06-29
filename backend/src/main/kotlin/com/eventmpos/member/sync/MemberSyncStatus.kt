package com.eventmpos.member.sync

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table
import org.springframework.data.jpa.repository.JpaRepository
import java.time.Instant

@Entity
@Table(name = "member_sync_status")
class MemberSyncStatus(
    @Id val id: Int = 1,
    @Column(name = "last_sync_at") var lastSyncAt: Instant? = null,
    @Column(name = "last_result") var lastResult: String? = null,
    @Column(name = "member_count") var memberCount: Int = 0,
    var source: String? = null,
    var message: String? = null,
)

interface MemberSyncStatusRepository : JpaRepository<MemberSyncStatus, Int>
