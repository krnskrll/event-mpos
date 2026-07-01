package com.eventmpos.report

import com.eventmpos.audit.AuditEntryResponse
import com.eventmpos.audit.AuditService
import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/**
 * Admin-only reporting + audit surface (under /api/admin, locked to ROLE_ADMIN in SecurityConfig).
 * - end-of-event report (JSON + CSV export)
 * - append-only order audit trail for an event (or a single order)
 */
@RestController
@RequestMapping("/api/admin")
class ReportAdminController(
    private val reportService: ReportService,
    private val auditService: AuditService,
) {
    @GetMapping("/events/{eventId}/report")
    fun report(@PathVariable eventId: UUID): EndOfEventReport = reportService.endOfEvent(eventId)

    @GetMapping("/events/{eventId}/report.csv", produces = ["text/csv"])
    fun reportCsv(@PathVariable eventId: UUID): ResponseEntity<String> {
        val report = reportService.endOfEvent(eventId)
        val csv = reportService.toCsv(report)
        val filename = "report-${report.eventCode}.csv"
        return ResponseEntity.ok()
            .header(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=\"$filename\"")
            .contentType(MediaType.parseMediaType("text/csv"))
            .body(csv)
    }

    @GetMapping("/events/{eventId}/audit")
    fun audit(
        @PathVariable eventId: UUID,
        @RequestParam(required = false) orderId: UUID?,
    ): List<AuditEntryResponse> =
        if (orderId != null) auditService.forOrder(orderId) else auditService.forEvent(eventId)
}
