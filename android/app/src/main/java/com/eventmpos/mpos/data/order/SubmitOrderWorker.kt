package com.eventmpos.mpos.data.order

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject

/**
 * Drains PENDING orders to the server via [OrderSubmitter]. Retried by WorkManager (network
 * constraint + backoff) until every order is CONFIRMED or permanently FAILED.
 */
@HiltWorker
class SubmitOrderWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted params: WorkerParameters,
    private val submitter: OrderSubmitter,
) : CoroutineWorker(appContext, params) {

    override suspend fun doWork(): Result =
        if (submitter.drainPending()) Result.retry() else Result.success()
}
