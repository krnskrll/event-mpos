package com.eventmpos.mpos.data.order

import com.eventmpos.mpos.data.auth.AuthApi
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.IOException

class OrderNumbersTest {
    /** In-memory blocks with the same "lowest unfinished block first" rule as the Room queries. */
    private class FakeDao : NumberBlockDao() {
        val blocks = mutableListOf<NumberBlockEntity>()
        private fun open(e: String) = blocks.filter { it.eventId == e && it.nextUnused <= it.rangeEnd }.sortedBy { it.rangeStart }
        override suspend fun current(eventId: String) = open(eventId).firstOrNull()
        override suspend fun remaining(eventId: String) = open(eventId).sumOf { it.rangeEnd - it.nextUnused + 1 }
        override suspend fun insert(block: NumberBlockEntity) { blocks += block }
        override suspend fun advance(eventId: String, rangeStart: Int) {
            val i = blocks.indexOfFirst { it.eventId == eventId && it.rangeStart == rangeStart }
            blocks[i] = blocks[i].copy(nextUnused = blocks[i].nextUnused + 1)
        }
    }

    private val dao = FakeDao()
    private val api = mockk<AuthApi>()
    private val numbers = OrderNumbers(dao, api, mockk { coEvery { id() } returns "device-1" })

    @Test
    fun `hands out each number once, across blocks, then blocks checkout`() = runTest {
        dao.insert(NumberBlockEntity("ev1", "20261030", 1, 2, 1))
        dao.insert(NumberBlockEntity("ev1", "20261030", 52, 52, 52))
        assertEquals(listOf(1, 2, 52), List(3) { numbers.next("ev1").first })
        assertTrue(runCatching { numbers.next("ev1") }.exceptionOrNull() is NoOrderNumberException)
    }

    @Test
    fun `tops up below ten remaining and tolerates being offline`() = runTest {
        coEvery { api.issueNumberBlock("ev1", NumberBlockRequestDto("device-1")) } returns NumberBlockDto("20261030", 1, 50)
        numbers.topUp("ev1")
        numbers.topUp("ev1") // 50 left: no second request
        coVerify(exactly = 1) { api.issueNumberBlock(any(), any()) }
        assertEquals(1 to "20261030", numbers.next("ev1"))

        coEvery { api.issueNumberBlock(any(), any()) } throws IOException("offline")
        repeat(45) { numbers.next("ev1") } // 4 left
        numbers.topUp("ev1") // offline: no crash, still 4 to use
        assertEquals(4, dao.remaining("ev1"))
    }
}
