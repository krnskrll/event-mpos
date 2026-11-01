package com.eventmpos.mpos

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.Build
import com.sunmi.peripheral.printer.InnerPrinterCallback
import com.sunmi.peripheral.printer.InnerPrinterManager
import com.sunmi.peripheral.printer.SunmiPrinterService
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.EventChannel
import io.flutter.plugin.common.MethodChannel

/** Sunmi built-in printer and 2D scanner for the Flutter side (lib/core/printer/sunmi.dart). */
class MainActivity : FlutterActivity() {
    private val isSunmi = Build.MANUFACTURER.equals("SUNMI", ignoreCase = true)
    private var printer: SunmiPrinterService? = null

    override fun configureFlutterEngine(engine: FlutterEngine) {
        super.configureFlutterEngine(engine)
        val messenger = engine.dartExecutor.binaryMessenger
        MethodChannel(messenger, "mpos/sunmi").setMethodCallHandler { call, result ->
            when (call.method) {
                "isSunmi" -> result.success(isSunmi)
                "status" -> result.success(status())
                "print" -> result.success(print(call.argument<List<String>>("lines").orEmpty()))
                else -> result.notImplemented()
            }
        }
        EventChannel(messenger, "mpos/sunmi/scans").setStreamHandler(ScanReceiver(this))
        if (isSunmi) bindPrinter()
    }

    private fun bindPrinter() = runCatching {
        InnerPrinterManager.getInstance().bindService(this, object : InnerPrinterCallback() {
            override fun onConnected(service: SunmiPrinterService) { printer = service }
            override fun onDisconnected() { printer = null }
        })
    }

    private fun status(): Int = runCatching { printer?.updatePrinterState() }.getOrNull() ?: NO_PRINTER

    private fun print(lines: List<String>): Int {
        val service = printer ?: return NO_PRINTER
        val status = status()
        if (status != READY) return status
        return runCatching {
            lines.forEach { service.printText("$it\n", null) }
            service.lineWrap(3, null) // feed past the tear bar (the V2 Pro has no cutter)
            READY
        }.getOrDefault(NO_PRINTER)
    }

    private companion object {
        const val READY = 1
        const val NO_PRINTER = 505
    }
}

/** Forwards codes from the Sunmi scanner service, which broadcasts every read. */
private class ScanReceiver(private val context: Context) : EventChannel.StreamHandler {
    private var receiver: BroadcastReceiver? = null

    override fun onListen(arguments: Any?, sink: EventChannel.EventSink) {
        val r = object : BroadcastReceiver() {
            override fun onReceive(c: Context, intent: Intent) {
                intent.getStringExtra("data")?.let(sink::success)
            }
        }
        val filter = IntentFilter("com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED")
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            context.registerReceiver(r, filter, Context.RECEIVER_EXPORTED)
        } else {
            context.registerReceiver(r, filter)
        }
        receiver = r
    }

    override fun onCancel(arguments: Any?) {
        receiver?.let(context::unregisterReceiver)
        receiver = null
    }
}
