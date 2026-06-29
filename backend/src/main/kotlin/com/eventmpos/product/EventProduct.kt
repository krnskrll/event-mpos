package com.eventmpos.product

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import jakarta.persistence.UniqueConstraint
import java.util.UUID

/**
 * Marks a catalogue product as available for a specific event. Admins choose which products sell at
 * each event; the Android snapshot only includes products that are available for that event.
 */
@Entity
@Table(
    name = "event_products",
    uniqueConstraints = [UniqueConstraint(name = "uk_event_product", columnNames = ["event_id", "product_id"])],
    indexes = [Index(name = "idx_event_product_event", columnList = "event_id")],
)
class EventProduct(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "event_id", nullable = false)
    val eventId: UUID,

    @Column(name = "product_id", nullable = false)
    val productId: UUID,

    @Column(nullable = false)
    var available: Boolean = true,
)
