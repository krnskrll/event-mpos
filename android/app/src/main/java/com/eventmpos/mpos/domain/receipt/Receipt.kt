package com.eventmpos.mpos.domain.receipt

import com.eventmpos.mpos.core.format.rupiah

/** Everything needed to render + reprint a receipt, captured at confirm time (no network to reprint). */
data class ReceiptLine(val name: String, val qty: Int, val lineTotal: Double)

data class ReceiptData(
    val brand: String,
    val eventName: String,
    val orderNumberDisplay: String,   // "#042"
    val referenceId: String,
    val customerName: String,
    val customerTier: String?,
    val items: List<ReceiptLine>,
    val subtotal: Double,
    val discount: Double,
    val promotionName: String?,
    val total: Double,
    val cashierNote: String,
    val dateText: String,
)

/**
 * Renders a [ReceiptData] into 58mm thermal lines (32 chars at the standard font). Pure + deterministic
 * so it's unit-testable and identical whatever the PrinterPort. Prices are right-aligned into columns.
 */
object ReceiptFormatter {
    const val WIDTH = 32

    fun format(r: ReceiptData): List<String> {
        val out = mutableListOf<String>()
        out += center(r.brand.uppercase())
        out += center(r.eventName)
        out += divider()
        out += lr("Order ${r.orderNumberDisplay}", r.dateText)
        out += r.referenceId
        out += "Customer: ${r.customerName}" + (r.customerTier?.let { " ($it)" } ?: "")
        out += divider()
        r.items.forEach { item ->
            // line 1: "qty x name", line 2: right-aligned amount (keeps long names readable on 58mm)
            out += "${item.qty} x ${item.name}".take(WIDTH)
            out += lr("", money(item.lineTotal))
        }
        out += divider()
        out += lr("Subtotal", money(r.subtotal))
        if (r.discount > 0.0) {
            out += lr("Discount" + (r.promotionName?.let { " ($it)" } ?: ""), "-" + money(r.discount))
        }
        out += lr("TOTAL", money(r.total))
        out += divider()
        wrap(r.cashierNote).forEach { out += center(it) }
        out += ""
        out += center("Payment at cashier counter")
        out += center("Not a tax invoice")
        // Safety: never exceed the 58mm width on any line.
        return out.map { if (it.length > WIDTH) it.take(WIDTH) else it }
    }

    private fun divider() = "-".repeat(WIDTH)

    /** Greedy word-wrap to the receipt width. */
    private fun wrap(text: String): List<String> {
        val lines = mutableListOf<String>()
        var line = StringBuilder()
        text.split(" ").forEach { word ->
            if (line.isEmpty()) line.append(word)
            else if (line.length + 1 + word.length <= WIDTH) line.append(' ').append(word)
            else { lines += line.toString(); line = StringBuilder(word) }
        }
        if (line.isNotEmpty()) lines += line.toString()
        return lines
    }

    private fun center(text: String): String {
        val t = text.take(WIDTH)
        val pad = (WIDTH - t.length) / 2
        return " ".repeat(pad.coerceAtLeast(0)) + t
    }

    /** Left text + right text on the same 32-char line, space-filled between. */
    private fun lr(left: String, right: String): String {
        val space = WIDTH - right.length
        val l = if (left.length > space) left.take((space - 1).coerceAtLeast(0)) else left
        val gap = (WIDTH - l.length - right.length).coerceAtLeast(1)
        return l + " ".repeat(gap) + right
    }

    private fun money(amount: Double) = rupiah(amount)
}
