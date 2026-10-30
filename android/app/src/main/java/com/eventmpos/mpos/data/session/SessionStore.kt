package com.eventmpos.mpos.data.session

import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.EventStatus
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Holds the current selling session — the BA's name and the event they selected — so the bottom-nav
 * tab screens (Home/Catalogue/Cart/Orders) share it without threading the eventId through routes.
 */
@Singleton
class SessionStore @Inject constructor() {
    private val _event = MutableStateFlow<AssignedEvent?>(null)
    val event: StateFlow<AssignedEvent?> = _event.asStateFlow()

    var userName: String = ""
        private set

    fun startSession(userName: String, event: AssignedEvent) {
        this.userName = userName
        _event.value = event
    }

    fun setUser(name: String) { userName = name }

    val eventId: String? get() = _event.value?.id

    /** Refresh the selected event's status (e.g. it was closed while the BA was selling). */
    fun updateStatus(status: EventStatus) { _event.value = _event.value?.copy(status = status) }

    fun clear() {
        userName = ""
        _event.value = null
    }
}
