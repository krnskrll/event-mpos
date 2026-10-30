package com.eventmpos.mpos.data.cart

import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.SelectedCustomer
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class CartStoreTest {

    private val store = CartStore()

    @Test
    fun `adding the same product accumulates quantity`() {
        store.add("p1")
        store.add("p1", 2)
        assertEquals(mapOf(CartKey.of("p1", null) to 3), store.items.value)
    }

    @Test
    fun `variants of one product are separate lines`() {
        store.addVariant("p1", "rose")
        store.addVariant("p1", "nude", 2)
        store.add("p1")
        assertEquals(3, store.items.value.size)
        assertEquals(2, store.items.value[CartKey.of("p1", "nude")])
    }

    @Test
    fun `stepping a line to zero removes it`() {
        store.add("p1", 2)
        val key = CartKey.of("p1", null)
        store.adjust(key, -1)
        assertEquals(1, store.items.value[key])
        store.adjust(key, -1)
        assertFalse(key in store.items.value)
        assertTrue(store.isEmpty())
    }

    @Test
    fun `remove drops the line regardless of quantity`() {
        store.add("p1", 5)
        store.add("p2")
        store.remove(CartKey.of("p1", null))
        assertEquals(setOf(CartKey.of("p2", null)), store.items.value.keys)
    }

    @Test
    fun `clear empties lines and customer`() {
        store.add("p1")
        store.setCustomer(SelectedCustomer("Aisha", "GOLD", "MBR-1", CustomerSource.entries.first()))
        store.clear()
        assertTrue(store.isEmpty())
        assertNull(store.customer.value)
    }

    @Test
    fun `cart key round-trips product and variant`() {
        val withVariant = CartKey.of("p1", "v9")
        assertEquals("p1", CartKey.productId(withVariant))
        assertEquals("v9", CartKey.variantId(withVariant))
        val base = CartKey.of("p1", null)
        assertEquals("p1", CartKey.productId(base))
        assertNull(CartKey.variantId(base))
    }
}
