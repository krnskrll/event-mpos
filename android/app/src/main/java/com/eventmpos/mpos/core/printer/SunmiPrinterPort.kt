package com.eventmpos.mpos.core.printer

/**
 * PRODUCTION printer for the Sunmi V2 Pro built-in 58mm thermal printer, via the Sunmi InnerPrinter
 * SDK over AIDL (NOT Bluetooth / network). Not wired by default (the app ships with FakePrinterPort
 * so it runs on emulators); on a Sunmi device, bind this in AppModule instead.
 *
 * Implementation notes (how the real binding works):
 *  1. Add the Sunmi printer AIDL to the project (woyou.aidlservice.jiuiv5.IWoyouService) — copy the
 *     vendor .aidl into app/src/main/aidl/woyou/aidlservice/jiuiv5/.
 *  2. Bind the service:
 *       val intent = Intent().apply {
 *         setPackage("woyou.aidlservice.jiuiv5")
 *         action = "woyou.aidlservice.jiuiv5.IWoyouService"
 *       }
 *       context.bindService(intent, connection, Context.BIND_AUTO_CREATE)
 *     and keep the IWoyouService from onServiceConnected.
 *  3. status(): query woyouService.updatePrinter / getPrinterStatus codes and map to PrinterStatus
 *     (e.g. 0=READY, out-of-paper, overheated, abnormal).
 *  4. print(lines): for each line call service.printText(line + "\n", callback); set alignment with
 *     setAlignment(); finish with lineWrap(3) + (optional) cutPaper; the V2 Pro has no cutter so feed
 *     a few lines instead. Use the ICallback to detect print failures.
 *
 * Until the vendor AIDL is present this throws if selected, by design.
 */
class SunmiPrinterPort : PrinterPort {
    override suspend fun status(): PrinterStatus = PrinterStatus.UNAVAILABLE
    override suspend fun print(lines: List<String>): PrintResult =
        PrintResult.Failure(PrinterStatus.UNAVAILABLE, "Sunmi InnerPrinter AIDL not bound in this build")
}
