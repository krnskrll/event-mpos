package com.eventmpos.event

import com.eventmpos.common.ApiException
import org.springframework.http.HttpStatus

/**
 * Event lifecycle. BAs can only sell while an event is ACTIVE.
 *
 * Allowed transitions:
 *   DRAFT  -> ACTIVE   (open the event for selling)
 *   ACTIVE -> CLOSED   (end the event)
 *   DRAFT  -> CLOSED   (cancel an event that never opened)
 * Anything else is rejected.
 */
enum class EventStatus {
    DRAFT,
    ACTIVE,
    CLOSED,
    ;

    fun assertCanTransitionTo(target: EventStatus) {
        val allowed = when (this) {
            DRAFT -> target == ACTIVE || target == CLOSED
            ACTIVE -> target == CLOSED
            CLOSED -> false
        }
        if (!allowed) {
            throw ApiException(
                HttpStatus.CONFLICT,
                "Cannot change event status from $this to $target",
            )
        }
    }
}
