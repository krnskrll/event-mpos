package com.eventmpos.mpos.core.format

import java.util.Locale
import kotlin.math.roundToLong

/**
 * Single source of currency formatting for the app. Indonesian Rupiah: "Rp" prefix, thousands
 * grouped with ".", no decimal places (Rupiah has no sub-unit in practice). e.g. 1032.0 -> "Rp 1.032".
 */
fun rupiah(amount: Double): String =
    "Rp " + String.format(Locale.US, "%,d", amount.roundToLong()).replace(',', '.')
