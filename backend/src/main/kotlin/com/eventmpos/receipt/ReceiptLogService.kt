package com.eventmpos.receipt

import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.time.Instant
import java.util.UUID

interface ReceiptLogRepository : JpaRepository<ReceiptLog, UUID> {
    fun findByOrderId(orderId: UUID): ReceiptLog?
}

/** Records print/reprint events. First print inserts the row; reprints bump the counter. */
@Service
class ReceiptLogService(private val repository: ReceiptLogRepository) {

    @Transactional
    fun logPrint(orderId: UUID) {
        val existing = repository.findByOrderId(orderId)
        if (existing == null) {
            repository.save(ReceiptLog(orderId = orderId))
        } else {
            existing.printCount += 1
            existing.lastPrintedAt = Instant.now()
            repository.save(existing)
        }
    }
}
