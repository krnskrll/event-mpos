package com.eventmpos.mpos.core.scanner

import kotlinx.coroutines.flow.Flow

/**
 * Hardware-agnostic barcode scanner boundary — the scanner equivalent of the PrinterPort rule in
 * .cursorrules. The selling flow depends only on this interface, so the real Sunmi 2D scanner (via
 * its AIDL service) can be swapped in without touching any UI/domain code, and tests/non-Sunmi
 * devices can use a manual-entry fallback.
 *
 * The real implementation (SunmiScannerPort) binds the Sunmi Scanner AIDL service and emits decoded
 * values on [scans]; it is added when building on-device against the Sunmi SDK. The default build
 * uses [ManualScannerPort] so the app runs on any emulator.
 */
interface ScannerPort {
    /** Stream of decoded barcode strings as they are scanned. */
    val scans: Flow<String>

    /** Feed a barcode manually (typed entry, or a test/emulator). */
    suspend fun emitManual(barcode: String)
}
