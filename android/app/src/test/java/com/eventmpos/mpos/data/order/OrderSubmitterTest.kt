package com.eventmpos.mpos.data.order

import com.eventmpos.mpos.core.notification.OrderNotifier
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.PendingOrderEntity
import com.eventmpos.mpos.data.receipt.ReceiptManager
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.coVerifyOrder
import io.mockk.mockk
import io.mockk.slot
import io.mockk.verify
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import retrofit2.HttpException
import retrofit2.Response
import java.io.IOException

/** PENDING → CONFIRMED / FAILED transitions of the upload queue (the worker's logic). */
class OrderSubmitterTest {

    private val api = mockk<AuthApi>()
    private val dao = FakePendingOrderDao()
    private val receipts = mockk<ReceiptManager>(relaxed = true)
    private val notifier = mockk<OrderNotifier>(relaxed = true)
    private val submitter = OrderSubmitter(api, dao, receipts, notifier, Json { ignoreUnknownKeys = true })

    private suspend fun queue(key: String, at: Long = 1) = dao.upsert(
        PendingOrderEntity(
            idempotencyKey = key, eventId = "ev1", customerName = "Cust $key", customerQr = null,
            customerTier = null, linesJson = """[{"productId":"p1","qty":1}]""", status = "PENDING",
            createdAtEpochMs = at,
        ),
    )

    private fun confirmation(n: Int) = OrderConfirmationDto(
        id = "srv-$n", orderNumber = n, orderNumberDisplay = "#%03d".format(n),
        referenceId = "ORD-20261030-EV01-%03d".format(n), status = "NEW", customerName = "x", total = 49_500.0,
    )

    private fun httpError(code: Int) = HttpException(Response.error<Any>(code, "{}".toResponseBody()))

    @Test
    fun `server confirmation marks CONFIRMED, prints and notifies`() = runTest {
        queue("k1")
        val sent = slot<SubmitOrderRequestDto>()
        coEvery { api.submitOrder("ev1", capture(sent)) } returns confirmation(7)

        assertFalse(submitter.drainPending())

        val row = dao.get("k1")!!
        assertEquals("CONFIRMED", row.status)
        assertEquals("#007", row.orderNumberDisplay)
        assertEquals("ORD-20261030-EV01-007", row.referenceId)
        assertEquals("k1", sent.captured.idempotencyKey) // the stored key is reused on every attempt
        assertEquals(listOf(OrderLineDto("p1", 1)), sent.captured.lines)
        coVerify { receipts.storeAndPrint("k1", "ev1", any()) }
        verify { notifier.notifyConfirmed("k1", "#007", 49_500.0) }
    }

    @Test
    fun `network failure keeps the order PENDING and asks for a retry`() = runTest {
        queue("k1")
        coEvery { api.submitOrder(any(), any()) } throws IOException("offline")

        assertTrue(submitter.drainPending())
        assertEquals("PENDING", dao.get("k1")!!.status)
        coVerify(exactly = 0) { receipts.storeAndPrint(any(), any(), any()) }
    }

    @Test
    fun `server error 5xx keeps the order PENDING`() = runTest {
        queue("k1")
        coEvery { api.submitOrder(any(), any()) } throws httpError(503)

        assertTrue(submitter.drainPending())
        assertEquals("PENDING", dao.get("k1")!!.status)
    }

    @Test
    fun `client error 4xx marks FAILED permanently`() = runTest {
        queue("k1")
        coEvery { api.submitOrder(any(), any()) } throws httpError(409)

        assertFalse(submitter.drainPending())
        val row = dao.get("k1")!!
        assertEquals("FAILED", row.status)
        assertEquals("Rejected (409)", row.errorMessage)
        verify { notifier.notifyFailed("k1", "Cust k1", "Rejected (409)") }
        assertTrue(dao.pending().isEmpty()) // never retried again
    }

    @Test
    fun `a print failure does not affect the confirmed order`() = runTest {
        queue("k1")
        coEvery { api.submitOrder(any(), any()) } returns confirmation(1)
        coEvery { receipts.storeAndPrint(any(), any(), any()) } throws IllegalStateException("printer on fire")

        assertFalse(submitter.drainPending())
        assertEquals("CONFIRMED", dao.get("k1")!!.status)
    }

    @Test
    fun `queue drains oldest first and one failure does not block the rest`() = runTest {
        queue("late", at = 30)
        queue("early", at = 10)
        queue("middle", at = 20)
        coEvery { api.submitOrder(any(), match { it.idempotencyKey == "middle" }) } throws IOException("blip")
        coEvery { api.submitOrder(any(), match { it.idempotencyKey != "middle" }) } returns confirmation(1)

        assertTrue(submitter.drainPending())

        coVerifyOrder {
            api.submitOrder(any(), match { it.idempotencyKey == "early" })
            api.submitOrder(any(), match { it.idempotencyKey == "middle" })
            api.submitOrder(any(), match { it.idempotencyKey == "late" })
        }
        assertEquals("CONFIRMED", dao.get("early")!!.status)
        assertEquals("PENDING", dao.get("middle")!!.status)
        assertEquals("CONFIRMED", dao.get("late")!!.status)
    }
}
