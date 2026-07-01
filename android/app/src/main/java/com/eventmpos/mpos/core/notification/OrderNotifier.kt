package com.eventmpos.mpos.core.notification

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Posts a local notification when a queued (offline) order finally resolves, so the BA knows a
 * "Pending" sale went through (or needs attention) without watching the Orders screen. A no-op if
 * the POST_NOTIFICATIONS permission was declined — the Orders list still shows the status.
 */
@Singleton
class OrderNotifier @Inject constructor(@ApplicationContext private val context: Context) {

    init {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(CHANNEL_ID, "Order updates", NotificationManager.IMPORTANCE_DEFAULT).apply {
                description = "Tells you when a queued order is confirmed or fails."
            }
            context.getSystemService(NotificationManager::class.java)?.createNotificationChannel(channel)
        }
    }

    fun notifyConfirmed(idempotencyKey: String, orderNumber: String?, total: Double?) {
        val subtitle = buildString {
            orderNumber?.let { append(it) }
            total?.let { if (isNotEmpty()) append(" · "); append("RM ${"%.2f".format(it)}") }
        }.ifBlank { "Submitted successfully" }
        post(idempotencyKey, "Order confirmed", subtitle)
    }

    fun notifyFailed(idempotencyKey: String, customerName: String, reason: String) {
        post(idempotencyKey, "Order failed — $customerName", reason)
    }

    private fun post(idempotencyKey: String, title: String, text: String) {
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.stat_sys_upload_done)
            .setContentTitle(title)
            .setContentText(text)
            .setAutoCancel(true)
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .build()
        // notify() throws if POST_NOTIFICATIONS was denied (API 33+); swallow — it's non-essential.
        runCatching { NotificationManagerCompat.from(context).notify(idempotencyKey.hashCode(), notification) }
    }

    private companion object {
        const val CHANNEL_ID = "order_updates"
    }
}
