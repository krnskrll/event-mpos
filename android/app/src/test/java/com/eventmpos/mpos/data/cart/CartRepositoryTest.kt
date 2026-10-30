package com.eventmpos.mpos.data.cart

import com.eventmpos.mpos.data.local.CartDraftDao
import com.eventmpos.mpos.data.local.CartDraftEntity
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.SelectedCustomer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class CartRepositoryTest {

    /** In-memory DAO; shared across "process restarts" in the autosave test. */
    private class FakeDao : CartDraftDao {
        val rows = MutableStateFlow<Map<String, CartDraftEntity>>(emptyMap())
        override fun observeDrafts(eventId: String) =
            rows.map { m -> m.values.filter { it.eventId == eventId && !it.live }.sortedByDescending { it.createdAtEpochMs } }
        override suspend fun get(id: String) = rows.value[id]
        override suspend fun upsert(draft: CartDraftEntity) { rows.value = rows.value + (draft.id to draft) }
        override suspend fun delete(id: String) { rows.value = rows.value - id }
        override suspend fun deleteForEvent(eventId: String) { rows.value = rows.value.filterValues { it.eventId != eventId } }
    }

    private val dao = FakeDao()
    private val aisha = SelectedCustomer("Aisha", "GOLD", "MBR-1", CustomerSource.LOCAL_CACHE)
    private fun event(id: String) = AssignedEvent(id, id, id, null, null, EventStatus.ACTIVE)

    /** Let background work (incl. the autosave debounce) run; advanceUntilIdle skips backgroundScope delays. */
    private fun TestScope.settle() { advanceTimeBy(1_000); runCurrent() }

    /** A fresh app process: new in-memory cart + session, same Room tables. */
    private fun TestScope.launchApp(): Triple<CartStore, SessionStore, CartRepository> {
        val cart = CartStore()
        val session = SessionStore()
        return Triple(cart, session, CartRepository(cart, dao, session, Json, backgroundScope))
    }

    @Test
    fun `save parks the cart and resume restores lines and customer`() = runTest {
        val (cart, session, repo) = launchApp()
        session.startSession("BA", event("ev1")); settle()
        cart.add("p1", 2); cart.addVariant("p2", "rose"); cart.setCustomer(aisha)

        assertTrue(repo.saveDraft("Aisha's basket"))
        assertTrue(cart.isEmpty())
        val draft = repo.observeDrafts("ev1").first().single()
        assertEquals(3, draft.itemCount)

        repo.resume(draft.id)
        assertEquals(mapOf(CartKey.of("p1", null) to 2, CartKey.of("p2", "rose") to 1), cart.items.value)
        assertEquals(aisha, cart.customer.value)
        assertTrue(repo.observeDrafts("ev1").first().isEmpty())
    }

    @Test
    fun `live cart survives the process being killed`() = runTest {
        val (cart, session) = launchApp()
        session.startSession("BA", event("ev1")); settle()
        cart.add("p1", 3); cart.setCustomer(aisha); settle()

        val (cart2, session2) = launchApp() // process restart
        session2.startSession("BA", event("ev1")); settle()
        assertEquals(mapOf(CartKey.of("p1", null) to 3), cart2.items.value)
        assertEquals(aisha, cart2.customer.value)
    }

    @Test
    fun `switching event drops the old event's drafts and cart`() = runTest {
        val (cart, session, repo) = launchApp()
        session.startSession("BA", event("ev1")); settle()
        cart.add("p1"); repo.saveDraft("old"); cart.add("p2"); settle()

        session.startSession("BA", event("ev2")); settle()
        assertTrue(repo.observeDrafts("ev1").first().isEmpty())
        assertTrue(cart.isEmpty())
        assertTrue(dao.rows.value.isEmpty())
    }
}
