package com.eventmpos.mpos.ui.session

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.event.EventRepository
import com.eventmpos.mpos.domain.event.SyncResult
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SyncUiState(
    val stepIndex: Int = 0,
    val progress: Float = 0f,
    val done: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class SyncViewModel @Inject constructor(
    private val eventRepository: EventRepository,
    private val sessionStore: SessionStore,
) : ViewModel() {

    val eventName: String get() = sessionStore.event.value?.name ?: "your event"

    val steps = listOf(
        "Downloading products…",
        "Syncing member data…",
        "Loading promotions…",
        "Fetching inventory…",
        "Finalising setup…",
    )

    private val _state = MutableStateFlow(SyncUiState())
    val state: StateFlow<SyncUiState> = _state.asStateFlow()

    init { run() }

    fun run() {
        _state.value = SyncUiState()
        viewModelScope.launch {
            // Kick the real snapshot sync in parallel with a stepped progress animation.
            val eventId = sessionStore.eventId
            val sync = async { eventId?.let { eventRepository.syncSnapshot(it) } ?: SyncResult.Failed("No event selected") }

            val perStep = 480L
            steps.indices.forEach { i ->
                _state.update { it.copy(stepIndex = i, progress = (i + 1f) / steps.size) }
                delay(perStep)
            }
            when (val result = sync.await()) {
                is SyncResult.Ready, is SyncResult.NotActive ->
                    _state.update { it.copy(progress = 1f, done = true) }
                is SyncResult.Failed ->
                    _state.update { it.copy(error = result.message) }
            }
        }
    }
}
