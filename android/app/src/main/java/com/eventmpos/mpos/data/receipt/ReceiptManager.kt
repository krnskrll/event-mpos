package com.eventmpos.mpos.data.receipt

import com.eventmpos.mpos.BuildConfig
import com.eventmpos.mpos.core.printer.PrintResult
import com.eventmpos.mpos.core.printer.PrinterPort
import com.eventmpos.mpos.core.printer.PrinterStatus
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.ReceiptDao
import com.eventmpos.mpos.data.local.ReceiptEntity
import com.eventmpos.mpos.data.order.OrderConfirmationDto
import com.eventmpos.mpos.domain.receipt.ReceiptData
import com.eventmpos.mpos.domain.receipt.ReceiptFormatter
import com.eventmpos.mpos.domain.receipt.ReceiptLine
import com.eventmpos.mpos.domain.receipt.ReceiptRepository
import com.eventmpos.mpos.domain.receipt.StoredReceipt
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Builds + stores receipts and drives the printer. Used by the order worker (auto-print on confirm)
 * and the receipt screen (reprint). The receipt is persisted BEFORE/ regardless of print success, so
 * a printer failure never loses it — the BA can reprint. The order itself is never affected here.
 */
@Singleton
class ReceiptManager @Inject constructor(
    private val dao: ReceiptDao,
    private val printer: PrinterPort,
    private val api: AuthApi,
) : ReceiptRepository {

    private val timeFmt = DateTimeFormatter.ofPattern("dd/MM HH:mm").withZone(ZoneId.systemDefault())

    /** Called by the worker the moment the server confirms an order: store, then auto-print. */
    suspend fun storeAndPrint(orderKey: String, eventId: String, c: OrderConfirmationDto): PrintResult {
        val data = ReceiptData(
            brand = BuildConfig.BRAND_NAME,
            eventName = c.eventName,
            orderNumberDisplay = c.orderNumberDisplay,
            referenceId = c.referenceId,
            customerName = c.customerName,
            customerTier = c.customerTier,
            items = c.items.map { ReceiptLine(it.name, it.qty, it.lineTotal) },
            subtotal = c.subtotal,
            discount = c.discount,
            promotionName = c.promotionName,
            total = c.total,
            cashierNote = c.cashierNote,
            dateText = runCatching { timeFmt.format(Instant.parse(c.createdAt)) }.getOrDefault(""),
        )
        val lines = ReceiptFormatter.format(data)
        dao.upsert(
            ReceiptEntity(
                orderKey = orderKey, eventId = eventId, orderNumberDisplay = c.orderNumberDisplay,
                receiptText = lines.joinToString("\n"), printedOk = false,
                createdAtEpochMs = System.currentTimeMillis(), backendOrderId = c.id,
            ),
        )
        val result = printStored(orderKey, lines)
        logReceiptToServer(c.id) // fire-and-forget audit; never blocks the print flow
        return result
    }

    override suspend fun get(orderKey: String): StoredReceipt? =
        dao.get(orderKey)?.let { StoredReceipt(it.orderKey, it.receiptText.split("\n"), it.printedOk) }

    override suspend fun reprint(orderKey: String): PrintResult {
        val stored = dao.get(orderKey) ?: return PrintResult.Failure(PrinterStatus.ERROR, "No stored receipt")
        val result = printStored(orderKey, stored.receiptText.split("\n"))
        stored.backendOrderId?.let { logReceiptToServer(it) } // bumps the server reprint counter
        return result
    }

    /** Best-effort: tell the backend a receipt was printed. Offline/failure is fine — never throws. */
    private suspend fun logReceiptToServer(backendOrderId: String) {
        runCatching { api.logReceipt(backendOrderId) }
    }

    private suspend fun printStored(orderKey: String, lines: List<String>): PrintResult {
        // status check before printing (paper/temperature/connection)
        val status = printer.status()
        if (status != PrinterStatus.READY) {
            dao.setPrinted(orderKey, false)
            return PrintResult.Failure(status, "Printer not ready: $status")
        }
        val result = printer.print(lines)
        dao.setPrinted(orderKey, result is PrintResult.Success)
        return result
    }
}
