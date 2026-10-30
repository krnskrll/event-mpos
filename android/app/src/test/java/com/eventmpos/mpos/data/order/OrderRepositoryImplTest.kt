package com.eventmpos.mpos.data.order

import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequest
import androidx.work.WorkManager
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.data.local.PendingOrderEntity
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.SelectedCustomer
import com.eventmpos.mpos.domain.order.OrderState
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class OrderRepositoryImplTest {

    private val cart = CartStore()
    private val dao = FakePendingOrderDao()
    private val workManager = mockk<WorkManager>(relaxed = true)
    private val json = Json { ignoreUnknownKeys = true }
    private val repo = OrderRepositoryImpl(cart, dao, workManager, json)

    private fun customer() = SelectedCustomer("Bryan Tan", "GOLD", "MBR-00002", CustomerSource.entries.first())

    @Test
    fun `confirming queues a PENDING order, enqueues the upload and clears the cart`() = runTest {
        cart.add("p1", 2)
        cart.addVariant("p2", "rose")
        cart.setCustomer(customer())

        val key = repo.confirmCurrentCart("ev1")

        val row = dao.get(key)!!
        assertEquals("PENDING", row.status)
        assertEquals("ev1", row.eventId)
        assertEquals("Bryan Tan", row.customerName)
        assertEquals("MBR-00002", row.customerQr)
        val lines = json.decodeFromString(ListSerializer(OrderLineDto.serializer()), row.linesJson)
        assertEquals(setOf(OrderLineDto("p1", 2, null), OrderLineDto("p2", 1, "rose")), lines.toSet())
        verify { workManager.enqueueUniqueWork("submit-orders", ExistingWorkPolicy.APPEND_OR_REPLACE, any<OneTimeWorkRequest>()) }
        assertTrue(cart.isEmpty())
        assertNull(cart.customer.value)
    }

    @Test
    fun `each confirm gets its own idempotency key`() = runTest {
        cart.add("p1"); cart.setCustomer(customer())
        val first = repo.confirmCurrentCart("ev1")
        cart.add("p1"); cart.setCustomer(customer())
        val second = repo.confirmCurrentCart("ev1")
        assertTrue(first != second)
        assertEquals(2, dao.pending().size)
    }

    @Test(expected = IllegalStateException::class)
    fun `confirming without a customer is refused`() = runTest {
        cart.add("p1")
        repo.confirmCurrentCart("ev1")
    }

    @Test(expected = IllegalArgumentException::class)
    fun `confirming an empty cart is refused`() = runTest {
        cart.setCustomer(customer())
        repo.confirmCurrentCart("ev1")
    }

    @Test
    fun `observeOrders maps every state and treats unknown status as pending`() = runTest {
        fun row(key: String, status: String, at: Long) = PendingOrderEntity(
            idempotencyKey = key, eventId = "ev1", customerName = key, customerQr = null, customerTier = null,
            linesJson = "[]", status = status, createdAtEpochMs = at,
        )
        dao.upsert(row("a", "PENDING", 1))
        dao.upsert(row("b", "CONFIRMED", 2).copy(orderNumberDisplay = "#007", referenceId = "ORD-1", total = 49_500.0))
        dao.upsert(row("c", "FAILED", 3).copy(errorMessage = "Rejected (409)"))
        dao.upsert(row("d", "SOMETHING_NEW", 4))

        val orders = repo.observeOrders("ev1").first().associateBy { it.idempotencyKey }
        assertEquals(OrderState.PENDING, orders["a"]!!.state)
        assertEquals(OrderState.CONFIRMED, orders["b"]!!.state)
        assertEquals("#007", orders["b"]!!.orderNumberDisplay)
        assertEquals(OrderState.FAILED, orders["c"]!!.state)
        assertEquals("Rejected (409)", orders["c"]!!.error)
        assertEquals(OrderState.PENDING, orders["d"]!!.state)
    }
}
