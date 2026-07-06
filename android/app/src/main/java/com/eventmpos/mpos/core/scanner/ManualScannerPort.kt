package com.eventmpos.mpos.core.scanner

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Default [ScannerPort]: barcodes come from manual entry (or a test). Lets the app run anywhere.
 * On a Sunmi device this is replaced by a binding to the Sunmi Scanner AIDL service.
 */
@Singleton
class ManualScannerPort @Inject constructor() : ScannerPort {
    private val _scans = MutableSharedFlow<String>(extraBufferCapacity = 16)
    override val scans = _scans.asSharedFlow()

    override suspend fun emitManual(barcode: String) {
        _scans.emit(barcode.trim())
    }
}
