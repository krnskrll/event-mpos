package com.eventmpos.product.erp

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table
import org.springframework.data.jpa.repository.JpaRepository
import java.time.Instant

/**
 * Single-row record of the most recent ERP sync, so admins can see "last synced at / result".
 * The fixed id keeps it a singleton.
 */
@Entity
@Table(name = "erp_sync_status")
class ErpSyncStatus(
    @Id
    val id: Int = 1,

    @Column(name = "last_sync_at")
    var lastSyncAt: Instant? = null,

    @Column(name = "last_result")
    var lastResult: String? = null,

    @Column(name = "product_count")
    var productCount: Int = 0,

    var source: String? = null,

    var message: String? = null,
)

interface ErpSyncStatusRepository : JpaRepository<ErpSyncStatus, Int>
