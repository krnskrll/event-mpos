package com.eventmpos.mpos.domain.receipt

import com.eventmpos.mpos.core.printer.PrintResult

/** Stored receipt as the device knows it (for the receipt screen / reprint). */
data class StoredReceipt(val orderKey: String, val lines: List<String>, val printedOk: Boolean)

interface ReceiptRepository {
    /** The receipt for an order, or null if none stored yet. */
    suspend fun get(orderKey: String): StoredReceipt?

    /** (Re)print an already-stored receipt — no new order, no network. Checks printer status first. */
    suspend fun reprint(orderKey: String): PrintResult
}
