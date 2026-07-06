package com.eventmpos.mpos.core.printer

import android.util.Log
import kotlinx.coroutines.delay
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Default [PrinterPort] for emulators / non-Sunmi devices: "prints" the receipt to Logcat (tag
 * `RECEIPT`) so the formatted 58mm output is visible during development, and reports READY.
 *
 * [simulateStatus] can be flipped (e.g. from a debug control) to exercise the failure + retry path
 * without real hardware.
 */
@Singleton
class FakePrinterPort @Inject constructor() : PrinterPort {

    @Volatile
    var simulateStatus: PrinterStatus = PrinterStatus.READY

    override suspend fun status(): PrinterStatus = simulateStatus

    override suspend fun print(lines: List<String>): PrintResult {
        val s = status()
        if (s != PrinterStatus.READY) {
            return PrintResult.Failure(s, "Printer not ready: $s")
        }
        delay(150) // mimic feed time
        Log.i("RECEIPT", "\n" + lines.joinToString("\n"))
        return PrintResult.Success
    }
}
