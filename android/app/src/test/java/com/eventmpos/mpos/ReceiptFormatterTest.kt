package com.eventmpos.mpos

import com.eventmpos.mpos.domain.receipt.ReceiptData
import com.eventmpos.mpos.domain.receipt.ReceiptFormatter
import com.eventmpos.mpos.domain.receipt.ReceiptLine
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ReceiptFormatterTest {

    private fun sample(discount: Double = 15.40) = ReceiptData(
        brand = "Lumière Beauty",
        eventName = "Mid-Year Beauty Roadshow",
        orderNumberDisplay = "#003",
        referenceId = "ORD-20260630-EV01-003",
        customerName = "Aisha Rahman",
        customerTier = "PLATINUM",
        items = listOf(
            ReceiptLine("Daily Mineral Sunscreen SPF50", 1, 99.00),
            ReceiptLine("Gentle Foaming Cleanser 150ml", 1, 55.00),
        ),
        subtotal = 154.00,
        discount = discount,
        promotionName = "10% off everything",
        total = 154.00 - discount,
        cashierNote = "Please proceed to the cashier to complete payment.",
        dateText = "30/06 13:24",
    )

    @Test
    fun `every line fits 58mm width`() {
        ReceiptFormatter.format(sample()).forEach {
            assertTrue("line too wide: '$it' (${it.length})", it.length <= ReceiptFormatter.WIDTH)
        }
    }

    @Test
    fun `includes the required receipt fields`() {
        val text = ReceiptFormatter.format(sample()).joinToString("\n")
        assertTrue(text.contains("LUMIÈRE BEAUTY"))             // brand (uppercased)
        assertTrue(text.contains("Mid-Year Beauty Roadshow"))   // event
        assertTrue(text.contains("#003"))                       // order number
        assertTrue(text.contains("Aisha Rahman"))               // customer
        assertTrue(text.contains("TOTAL"))
        assertTrue(text.contains("cashier"))                    // footer
        assertTrue(text.contains("Discount"))                   // discount line present
    }

    @Test
    fun `discount line omitted when no discount`() {
        val text = ReceiptFormatter.format(sample(discount = 0.0)).joinToString("\n")
        assertTrue(!text.contains("Discount"))
    }

    @Test
    fun `total reflects discount`() {
        val text = ReceiptFormatter.format(sample()).joinToString("\n")
        assertTrue(text.contains("Rp 139")) // 154.00 − 15.40 = 138.60, rendered as whole Rupiah
    }
}
