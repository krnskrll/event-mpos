package com.eventmpos.mpos.data.cart

import com.eventmpos.mpos.data.local.CartDraftDao
import com.eventmpos.mpos.data.local.CartDraftEntity
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.member.SelectedCustomer
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.drop
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

/** A parked cart, as listed in the drafts dialog. */
data class CartDraft(val id: String, val label: String, val itemCount: Int, val customerName: String?, val createdAtEpochMs: Long)

/**
 * Persists carts on the device: named drafts the BA parks and resumes, plus an autosaved copy of
 * the live cart so it survives the app being killed. Drafts belong to one event and are dropped
 * when the BA switches to another event or the event closes.
 */
@Singleton
class CartRepository @Inject constructor(
    private val cart: CartStore,
    private val dao: CartDraftDao,
    private val session: SessionStore,
    private val json: Json,
    scope: CoroutineScope,
) {
    init {
        scope.launch { followSession() }
    }

    fun observeDrafts(eventId: String): Flow<List<CartDraft>> = dao.observeDrafts(eventId).map { rows ->
        rows.map { CartDraft(it.id, it.label, linesOf(it).values.sum(), customerOf(it)?.name, it.createdAtEpochMs) }
    }

    /** Park the current cart as a named draft and start a fresh one. False if there is nothing to save. */
    suspend fun saveDraft(label: String): Boolean {
        val eventId = session.eventId ?: return false
        if (cart.isEmpty()) return false
        dao.upsert(entity(UUID.randomUUID().toString(), eventId, label.ifBlank { "Draft" }, live = false))
        cart.clear()
        return true
    }

    /** Load a draft's lines and customer into the cart (replacing it) and delete the draft. */
    suspend fun resume(id: String) {
        val draft = dao.get(id) ?: return
        cart.replace(linesOf(draft), customerOf(draft))
        dao.delete(id)
    }

    suspend fun delete(id: String) = dao.delete(id)

    /** Drop every draft and the autosaved cart of an event (event switched away from, or closed). */
    suspend fun clearEvent(eventId: String) {
        dao.deleteForEvent(eventId)
        if (session.eventId == eventId) cart.clear()
    }

    /**
     * Tracks the selected event. Entering an event restores its autosaved cart *before* autosave
     * starts (so an empty cart can't overwrite it); moving to a different event clears the old one.
     */
    @OptIn(FlowPreview::class)
    private suspend fun followSession() {
        var previous: String? = null
        session.event.map { it?.id }.distinctUntilChanged().collectLatest { eventId ->
            if (eventId == null) return@collectLatest // logged out: keep everything for next time
            previous?.takeIf { it != eventId }?.let { clearEvent(it); cart.clear() }
            previous = eventId
            if (cart.isEmpty()) dao.get(liveId(eventId))?.let { cart.replace(linesOf(it), customerOf(it)) }
            combine(cart.items, cart.customer) { items, customer -> items to customer }
                .drop(1).debounce(AUTOSAVE_DEBOUNCE_MS)
                .collect { (items, customer) ->
                    if (items.isEmpty() && customer == null) dao.delete(liveId(eventId))
                    else dao.upsert(entity(liveId(eventId), eventId, "", live = true))
                }
        }
    }

    private fun entity(id: String, eventId: String, label: String, live: Boolean) = CartDraftEntity(
        id = id, eventId = eventId, label = label, live = live, createdAtEpochMs = System.currentTimeMillis(),
        linesJson = json.encodeToString(cart.items.value),
        customerJson = cart.customer.value?.let { json.encodeToString(it) },
    )

    private fun linesOf(e: CartDraftEntity): Map<String, Int> = json.decodeFromString(e.linesJson)
    private fun customerOf(e: CartDraftEntity): SelectedCustomer? = e.customerJson?.let { json.decodeFromString(it) }
    private fun liveId(eventId: String) = "live:$eventId"

    private companion object {
        const val AUTOSAVE_DEBOUNCE_MS = 300L
    }
}
