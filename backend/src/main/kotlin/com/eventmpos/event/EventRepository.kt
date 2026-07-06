package com.eventmpos.event

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface EventRepository : JpaRepository<Event, UUID> {
    fun findByCode(code: String): Event?
    fun existsByCode(code: String): Boolean
}
