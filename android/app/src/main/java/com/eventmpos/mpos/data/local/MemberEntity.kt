package com.eventmpos.mpos.data.local

import androidx.room.Entity
import androidx.room.PrimaryKey

/**
 * Customer cached for instant offline QR identification. [qrCode] is the primary key (uniquely
 * indexed by Room) so a scan resolves in O(1) without network. Minimal fields only — display +
 * order linking, no financial data.
 */
@Entity(tableName = "member")
data class MemberEntity(
    @PrimaryKey val qrCode: String,
    val name: String,
    val tier: String,
)
