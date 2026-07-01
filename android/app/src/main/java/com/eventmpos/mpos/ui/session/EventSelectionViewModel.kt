package com.eventmpos.mpos.ui.session

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.session.SessionStore
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.event.EventRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class EventSelectionUiState(
    val loading: Boolean = true,
    val events: List<AssignedEvent> = emptyList(),
    val error: String? = null,
    val loggedOut: Boolean = false,
)

@HiltViewModel
class EventSelectionViewModel @Inject constructor(
    private val eventRepository: EventRepository,
    private val authRepository: AuthRepository,
    private val sessionStore: SessionStore,
) : ViewModel() {

    val userName: String get() = sessionStore.userName.ifBlank { "Beauty Advisor" }

    private val _state = MutableStateFlow(EventSelectionUiState())
    val state: StateFlow<EventSelectionUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        _state.update { it.copy(loading = true, error = null) }
        viewModelScope.launch {
            eventRepository.getAssignedEvents()
                .onSuccess { events -> _state.update { it.copy(loading = false, events = events) } }
                .onFailure { e -> _state.update { it.copy(loading = false, error = e.message ?: "Couldn't load your events") } }
        }
    }

    fun start(event: AssignedEvent) = sessionStore.startSession(userName, event)

    fun logout(onDone: () -> Unit) {
        viewModelScope.launch {
            authRepository.logout()
            sessionStore.clear()
            onDone()
        }
    }
}
