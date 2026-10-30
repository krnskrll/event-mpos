package com.eventmpos.mpos.data.receipt

import com.eventmpos.mpos.core.printer.PrintResult
import com.eventmpos.mpos.core.printer.PrinterPort
import com.eventmpos.mpos.core.printer.PrinterStatus
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.ReceiptDao
import com.eventmpos.mpos.data.local.ReceiptEntity
import com.eventmpos.mpos.domain.receipt.ReceiptData
import com.eventmpos.mpos.domain.receipt.ReceiptFormatter
import com.eventmpos.mpos.domain.receipt.ReceiptLine
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.IOException

class ReceiptManagerTest {

    private class FakeReceiptDao : ReceiptDao {
        val rows = mutableMapOf<String, ReceiptEntity>()
        override suspend fun upsert(receipt: ReceiptEntity) { rows[receipt.orderKey] = receipt }
        override suspend fun get(orderKey: String) = rows[orderKey]
        override suspend fun setPrinted(orderKey: String, ok: Boolean) {
            rows[orderKey]?.let { rows[orderKey] = it.copy(printedOk = ok) }
        }
    }

    private class FakePrinter(var status: PrinterStatus = PrinterStatus.READY, var fail: Boolean = false) : PrinterPort {
        val printed = mutableListOf<List<String>>()
        override suspend fun status() = status
        override suspend fun print(lines: List<String>): PrintResult {
            if (fail) return PrintResult.Failure(PrinterStatus.ERROR, "jam")
            printed += lines
            return PrintResult.Success
        }
    }

    private val dao = FakeReceiptDao()
    private val printer = FakePrinter()
    private val api = mockk<AuthApi>(relaxed = true)
    private val manager = ReceiptManager(dao, printer, api)

    private val data = ReceiptData(
        brand = "Lumière Beauty", eventName = "Mid-Year Beauty Roadshow", orderNumberDisplay = "#013",
        referenceId = "ORD-20261030-EV01-013", customerName = "Bryan Tan", customerTier = "GOLD",
        items = listOf(ReceiptLine("Gentle Foaming Cleanser 150ml", 1, 55_000.0)), subtotal = 55_000.0,
        discount = 5_500.0, promotionName = "10% off everything", total = 49_500.0,
        cashierNote = "Please proceed to the cashier to complete payment.", dateText = "30/10 14:31",
    )

    @Test
    fun `storeAndPrint works offline and linking after upload logs the print`() = runTest {
        val result = manager.storeAndPrint("k1", "ev1", data)

        assertEquals(PrintResult.Success, result)
        val stored = dao.rows["k1"]!!
        assertTrue(stored.printedOk && stored.receiptText.contains("Order #013") && stored.receiptText.contains("ORD-20261030-EV01-013"))
        assertEquals(1, printer.printed.size)
        coVerify(exactly = 0) { api.logReceipt(any()) } // nothing to log against until the order uploads

        manager.linkServerOrder("k1", "srv-13")
        assertEquals("srv-13", dao.rows["k1"]!!.backendOrderId)
        coVerify { api.logReceipt("srv-13") }
    }

    @Test
    fun `a cancelled order's receipt says so, once`() = runTest {
        manager.storeAndPrint("k1", "ev1", data)
        manager.markCancelled("k1"); manager.markCancelled("k1")
        val lines = dao.rows["k1"]!!.receiptText.split("\n")
        assertEquals(ReceiptFormatter.CANCELLED, lines[0])
        assertEquals(1, lines.count { it == ReceiptFormatter.CANCELLED })
    }

    @Test
    fun `printer not ready still stores the receipt for reprint`() = runTest {
        printer.status = PrinterStatus.OUT_OF_PAPER

        val result = manager.storeAndPrint("k1", "ev1", data)

        assertEquals(PrinterStatus.OUT_OF_PAPER, (result as PrintResult.Failure).status)
        assertNotNull(dao.rows["k1"])
        assertFalse(dao.rows["k1"]!!.printedOk)
        assertTrue(printer.printed.isEmpty())
    }

    @Test
    fun `a failed print can be reprinted once the printer recovers`() = runTest {
        printer.fail = true
        assertTrue(manager.storeAndPrint("k1", "ev1", data) is PrintResult.Failure)
        assertFalse(dao.rows["k1"]!!.printedOk)

        printer.fail = false
        assertEquals(PrintResult.Success, manager.reprint("k1"))
        assertTrue(dao.rows["k1"]!!.printedOk)
        assertEquals(dao.rows["k1"]!!.receiptText.split("\n"), printer.printed.single())
    }

    @Test
    fun `reprint works offline from the stored receipt`() = runTest {
        manager.storeAndPrint("k1", "ev1", data)
        coEvery { api.logReceipt(any()) } throws IOException("offline")

        assertEquals(PrintResult.Success, manager.reprint("k1"))
        assertEquals(2, printer.printed.size)
        assertEquals(printer.printed[0], printer.printed[1])
    }

    @Test
    fun `reprint without a stored receipt fails cleanly`() = runTest {
        val result = manager.reprint("missing")
        assertTrue(result is PrintResult.Failure)
        assertTrue(printer.printed.isEmpty())
    }

    @Test
    fun `server logging failure never breaks printing`() = runTest {
        coEvery { api.logReceipt(any()) } throws IOException("offline")
        assertEquals(PrintResult.Success, manager.storeAndPrint("k1", "ev1", data))
        assertTrue(dao.rows["k1"]!!.printedOk)
    }
}
