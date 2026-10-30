package com.eventmpos.mpos.data.order

import androidx.work.WorkManager
import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.local.ProductDao
import com.eventmpos.mpos.data.receipt.ReceiptManager
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.SelectedCustomer
import com.eventmpos.mpos.domain.order.OrderState
import com.eventmpos.mpos.domain.order.Sale
import com.eventmpos.mpos.domain.receipt.ReceiptData
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import io.mockk.verify
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import retrofit2.HttpException
import retrofit2.Response

/** Offline confirmation (EPIC-12 C1) and cancel (C4) on the device. */
class OrderRepositoryImplTest {
    private val cart = CartStore()
    private val dao = FakePendingOrderDao()
    private val numbers = mockk<OrderNumbers>(relaxed = true)
    private val receipts = mockk<ReceiptManager>(relaxed = true)
    private val products = mockk<ProductDao>(relaxed = true)
    private val api = mockk<AuthApi>(relaxed = true)
    private val workManager = mockk<WorkManager>(relaxed = true)
    private val session = SessionStore().apply {
        startSession("BA", AssignedEvent("ev1", "EV01", "Roadshow", null, null, EventStatus.ACTIVE))
    }
    private val sale = Sale(emptyList(), 55_000.0, 5_500.0, "10% off", 49_500.0)

    private fun TestScope.repo() =
        OrderRepositoryImpl(cart, dao, workManager, Json, numbers, receipts, products, session, api, backgroundScope)

    private fun fillCart() {
        cart.add("p1", 2)
        cart.setCustomer(SelectedCustomer("Bryan", "GOLD", "MBR-2", CustomerSource.LOCAL_CACHE))
    }

    @Test
    fun `confirm assigns a final number, prints, decrements local stock and queues the upload`() = runTest {
        coEvery { numbers.next("ev1") } returns (3 to "20261030")
        fillCart()

        val key = repo().confirmCurrentCart("ev1", sale)

        val row = dao.get(key)!!
        assertEquals(listOf("PENDING", "#003", "ORD-20261030-EV01-003"), listOf(row.status, row.orderNumberDisplay, row.referenceId))
        assertEquals(3, row.orderNumber)
        val printed = slot<ReceiptData>()
        coVerify { receipts.storeAndPrint(key, "ev1", capture(printed)) }
        assertEquals("#003", printed.captured.orderNumberDisplay)
        coVerify { products.adjustRemaining("ev1", "p1", -2) }
        verify { workManager.enqueueUniqueWork("submit-orders", any(), any<androidx.work.OneTimeWorkRequest>()) }
        assertTrue(cart.isEmpty())
    }

    @Test
    fun `no number left blocks checkout and keeps the cart`() = runTest {
        coEvery { numbers.next("ev1") } throws NoOrderNumberException()
        fillCart()

        val failure = runCatching { repo().confirmCurrentCart("ev1", sale) }.exceptionOrNull()

        assertTrue(failure is NoOrderNumberException)
        assertFalse(cart.isEmpty())
        assertTrue(dao.rows.value.isEmpty())
    }

    @Test
    fun `cancel marks the order and receipt cancelled and returns local stock`() = runTest {
        coEvery { numbers.next("ev1") } returns (1 to "20261030")
        fillCart()
        val r = repo()
        val key = r.confirmCurrentCart("ev1", sale)
        assertEquals("This order hasn't synced yet — try again when it's confirmed.", r.cancel(key))
        dao.upsert(dao.get(key)!!.copy(status = "CONFIRMED", serverOrderId = "srv-1"))
        assertTrue(r.observeOrders("ev1").first().single().canCancel)

        assertNull(r.cancel(key))
        assertEquals(OrderState.CANCELLED, r.observeOrders("ev1").first().single().state)
        coVerify { receipts.markCancelled(key) }
        coVerify { products.adjustRemaining("ev1", "p1", 2) }

        coEvery { api.cancelOrder(any(), any()) } throws HttpException(Response.error<Any>(409, "{}".toResponseBody()))
        dao.upsert(dao.get(key)!!.copy(status = "CONFIRMED"))
        assertTrue(r.cancel(key)!!.contains("already packing"))
    }
}
