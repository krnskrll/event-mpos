package com.eventmpos.mpos.data.receipt

import com.eventmpos.mpos.core.printer.PrintResult
import com.eventmpos.mpos.core.printer.PrinterPort
import com.eventmpos.mpos.core.printer.PrinterStatus
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.ReceiptDao
import com.eventmpos.mpos.data.local.ReceiptEntity
import com.eventmpos.mpos.domain.receipt.ReceiptData
import com.eventmpos.mpos.domain.receipt.ReceiptFormatter
import com.eventmpos.mpos.domain.receipt.ReceiptRepository
import com.eventmpos.mpos.domain.receipt.StoredReceipt
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Stores receipts and drives the printer. Used at confirmation (auto-print, works offline) and by the
 * receipt screen (reprint). The receipt is persisted BEFORE/ regardless of print success, so
 * a printer failure never loses it — the BA can reprint. The order itself is never affected here.
 */
@Singleton
class ReceiptManager @Inject constructor(
    private val dao: ReceiptDao,
    private val printer: PrinterPort,
    private val api: AuthApi,
) : ReceiptRepository {

    /** Called the moment the BA confirms (online or offline): store the receipt, then auto-print. */
    suspend fun storeAndPrint(orderKey: String, eventId: String, data: ReceiptData): PrintResult {
        val lines = ReceiptFormatter.format(data)
        dao.upsert(
            ReceiptEntity(
                orderKey = orderKey, eventId = eventId, orderNumberDisplay = data.orderNumberDisplay,
                receiptText = lines.joinToString("\n"), printedOk = false, createdAtEpochMs = System.currentTimeMillis(),
            ),
        )
        return printStored(orderKey, lines)
    }

    /** After upload: remember the server order id (reprints log against it) and log the first print. */
    suspend fun linkServerOrder(orderKey: String, backendOrderId: String) {
        val stored = dao.get(orderKey) ?: return
        dao.upsert(stored.copy(backendOrderId = backendOrderId))
        logReceiptToServer(backendOrderId)
    }

    /** Stamp the stored receipt CANCELLED, so the screen and any reprint show it. */
    suspend fun markCancelled(orderKey: String) {
        val stored = dao.get(orderKey) ?: return
        if (stored.receiptText.startsWith(ReceiptFormatter.CANCELLED)) return
        dao.upsert(stored.copy(receiptText = ReceiptFormatter.CANCELLED + "\n" + stored.receiptText))
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
