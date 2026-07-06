package com.eventmpos.sync

import org.springframework.data.jpa.repository.JpaRepository

interface SyncConfigRepository : JpaRepository<SyncConfig, String>
