package com.eventmpos.mpos.core.printer

/** Printer readiness, checked before every print attempt. */
enum class PrinterStatus { READY, OUT_OF_PAPER, OVERHEATED, ERROR, UNAVAILABLE }

sealed interface PrintResult {
    data object Success : PrintResult
    data class Failure(val status: PrinterStatus, val message: String) : PrintResult
}

/**
 * Hardware-agnostic printer boundary (the .cursorrules "PrinterPort abstraction"). All Sunmi
 * InnerPrinter SDK calls live behind this interface, so swapping hardware later means writing a new
 * implementation only — nothing in the receipt/print flow changes.
 *
 * Production impl: [SunmiPrinterPort] (binds the Sunmi InnerPrinter AIDL service on-device).
 * Default/dev impl: [FakePrinterPort] (renders to the log; lets the app run on any emulator).
 */
interface PrinterPort {
    /** Check the hardware before printing (paper/temperature/connection). */
    suspend fun status(): PrinterStatus

    /** Print a pre-formatted 58mm receipt (one string per line). */
    suspend fun print(lines: List<String>): PrintResult
}
