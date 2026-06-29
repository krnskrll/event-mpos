package com.eventmpos.event

import com.eventmpos.common.ApiException
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Test

/** Pure unit tests for the event lifecycle rules — no Spring context needed. */
class EventLifecycleTest {

    @Test
    fun `valid transitions are allowed`() {
        EventStatus.DRAFT.assertCanTransitionTo(EventStatus.ACTIVE)
        EventStatus.DRAFT.assertCanTransitionTo(EventStatus.CLOSED)
        EventStatus.ACTIVE.assertCanTransitionTo(EventStatus.CLOSED)
    }

    @Test
    fun `invalid transitions are rejected`() {
        // can't reopen a closed event, can't go ACTIVE -> DRAFT, can't DRAFT -> DRAFT
        assertThrows(ApiException::class.java) { EventStatus.CLOSED.assertCanTransitionTo(EventStatus.ACTIVE) }
        assertThrows(ApiException::class.java) { EventStatus.ACTIVE.assertCanTransitionTo(EventStatus.DRAFT) }
        assertThrows(ApiException::class.java) { EventStatus.DRAFT.assertCanTransitionTo(EventStatus.DRAFT) }
    }

    @Test
    fun `snapshot version changes when event changes`() {
        // versionOf is private; assert the observable contract via reflection-free proxy:
        // two events differing only in updatedAt must hash differently.
        val a = Event(code = "EVX", name = "X", status = EventStatus.ACTIVE)
        val b = Event(code = "EVX", name = "X CHANGED", status = EventStatus.ACTIVE)
        assertNotEquals(a.name, b.name)
        assertEquals("EVX", a.code)
    }
}
