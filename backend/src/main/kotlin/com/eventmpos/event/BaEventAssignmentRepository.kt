package com.eventmpos.event

import org.springframework.data.jpa.repository.JpaRepository
import java.util.UUID

interface BaEventAssignmentRepository : JpaRepository<BaEventAssignment, UUID> {
    fun findByUserIdAndActiveTrue(userId: UUID): BaEventAssignment?
    fun findAllByUserId(userId: UUID): List<BaEventAssignment>
    fun findAllByEventId(eventId: UUID): List<BaEventAssignment>
}
