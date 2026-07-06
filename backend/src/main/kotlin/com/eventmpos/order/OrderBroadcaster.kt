package com.eventmpos.order

import org.springframework.messaging.simp.SimpMessagingTemplate
import org.springframework.stereotype.Component
import org.springframework.transaction.event.TransactionPhase
import org.springframework.transaction.event.TransactionalEventListener
import java.util.UUID

/** Internal event fired when an order is created or advanced; carries the view to broadcast. */
data class OrderChangedEvent(val eventId: UUID, val view: WarehouseOrderView)

/**
 * Pushes order changes to the warehouse dashboard over STOMP. Broadcasts only AFTER the DB
 * transaction commits (so a rolled-back order never reaches the board), to
 * /topic/events/{eventId}/orders.
 */
@Component
class OrderBroadcaster(
    private val messagingTemplate: SimpMessagingTemplate,
) {
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    fun onCommitted(event: OrderChangedEvent) {
        messagingTemplate.convertAndSend("/topic/events/${event.eventId}/orders", event.view)
    }
}
