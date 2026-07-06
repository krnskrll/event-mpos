package com.eventmpos.mpos.core.connectivity

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * App-wide network reachability, exposed as a [StateFlow] the UI observes for the offline indicator
 * and the [com.eventmpos.mpos.core.sync.SyncCoordinator] uses to trigger a retry-queue flush on
 * reconnect. "Online" means the active network reports both INTERNET and VALIDATED capabilities, so
 * a captive/dead Wi-Fi at a venue reads as offline (which is what the BA experiences).
 */
@Singleton
class ConnectivityObserver @Inject constructor(@ApplicationContext context: Context) {

    private val cm = context.getSystemService(ConnectivityManager::class.java)
    private val _isOnline = MutableStateFlow(currentlyOnline())
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()

    init {
        cm?.registerDefaultNetworkCallback(object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) { recompute() }
            override fun onLost(network: Network) { _isOnline.value = false }
            override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
                _isOnline.value = caps.isInternet()
            }
        })
    }

    private fun recompute() { _isOnline.value = currentlyOnline() }

    private fun currentlyOnline(): Boolean {
        val active = cm?.activeNetwork ?: return false
        return cm.getNetworkCapabilities(active)?.isInternet() ?: false
    }

    private fun NetworkCapabilities.isInternet(): Boolean =
        hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) &&
            hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED)
}
