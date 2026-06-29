package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey

/**
 * A product cached for offline browsing/scanning. Scoped by eventId because availability is
 * per-event. Barcodes are stored as a newline-joined string (a product can have several);
 * scanning matches against this. Browsing reads from here, never the network (per .cursorrules).
 */
@Entity(
    tableName = "product",
    indices = [Index("eventId")],
)
data class ProductEntity(
    @PrimaryKey val rowKey: String, // "<eventId>:<id>" — unique per event+product
    val eventId: String,
    val id: String,
    val sku: String,
    val name: String,
    val brand: String?,
    val price: Double,
    val imageUrl: String?,
    val barcodes: String,
)
