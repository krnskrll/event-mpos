package com.eventmpos.promotion

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface PromotionRepository : JpaRepository<Promotion, UUID> {
    fun findAllByEventId(eventId: UUID): List<Promotion>
    fun findAllByEventIdAndActiveTrue(eventId: UUID): List<Promotion>
}
