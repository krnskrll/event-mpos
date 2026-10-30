package com.eventmpos.mpos

import android.app.Application
import androidx.hilt.work.HiltWorkerFactory
import androidx.work.Configuration
import com.eventmpos.mpos.data.cart.CartRepository
import dagger.hilt.android.HiltAndroidApp
import javax.inject.Inject

@HiltAndroidApp
class MposApplication : Application(), Configuration.Provider {

    @Inject
    lateinit var workerFactory: HiltWorkerFactory

    // Created eagerly so the live cart is restored/autosaved from the moment a session starts.
    @Inject
    lateinit var cartRepository: CartRepository

    // WorkManager picks up Hilt-injected workers (the order-submit retry queue).
    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder().setWorkerFactory(workerFactory).build()
}
