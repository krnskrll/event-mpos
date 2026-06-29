package com.eventmpos.mpos.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.domain.event.EventRepository
import com.eventmpos.mpos.domain.event.SyncResult
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

/** Sync sub-state for the selected event. */
sealed interface SyncUi {
    data object Idle : SyncUi
    data object Syncing : SyncUi
    data class Ready(val version: String, val unchanged: Boolean) : SyncUi
    data class Blocked(val message: String) : SyncUi
    data class Failed(val message: String) : SyncUi
}

data class HomeUiState(
    val loadingEvents: Boolean = true,
    val events: List<AssignedEvent> = emptyList(),
    val loadError: String? = null,
    val selected: AssignedEvent? = null,
    val sync: SyncUi = SyncUi.Idle,
    val loggedOut: Boolean = false,
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val eventRepository: EventRepository,
) : ViewModel() {

    private val _state = MutableStateFlow(HomeUiState())
    val state: StateFlow<HomeUiState> = _state.asStateFlow()

    init {
        loadEvents()
    }

    fun loadEvents() {
        _state.update { it.copy(loadingEvents = true, loadError = null) }
        viewModelScope.launch {
            eventRepository.getAssignedEvents()
                .onSuccess { events ->
                    _state.update { it.copy(loadingEvents = false, events = events) }
                    // If exactly one assigned event, select it automatically (EPIC-02 requirement).
                    if (events.size == 1) select(events.first())
                }
                .onFailure { err ->
                    _state.update {
                        it.copy(
                            loadingEvents = false,
                            loadError = err.message ?: "Couldn't load your events.",
                        )
                    }
                }
        }
    }

    fun select(event: AssignedEvent) {
        _state.update { it.copy(selected = event) }
        // Client-side lifecycle guard for instant feedback; the server enforces it too.
        if (event.status != EventStatus.ACTIVE) {
            _state.update {
                it.copy(sync = SyncUi.Blocked("Event ${event.code} is ${event.status} — selling opens when it's ACTIVE."))
            }
            return
        }
        sync(event)
    }

    fun retrySync() {
        _state.value.selected?.let { select(it) }
    }

    private fun sync(event: AssignedEvent) {
        _state.update { it.copy(sync = SyncUi.Syncing) }
        viewModelScope.launch {
            val result = eventRepository.syncSnapshot(event.id)
            _state.update {
                it.copy(
                    sync = when (result) {
                        is SyncResult.Ready -> SyncUi.Ready(result.version, result.unchanged)
                        is SyncResult.NotActive -> SyncUi.Blocked(result.message)
                        is SyncResult.Failed -> SyncUi.Failed(result.message)
                    },
                )
            }
        }
    }

    fun backToList() {
        _state.update { it.copy(selected = null, sync = SyncUi.Idle) }
    }

    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
            _state.update { it.copy(loggedOut = true) }
        }
    }
}
