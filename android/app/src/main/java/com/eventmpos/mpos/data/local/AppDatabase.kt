package com.eventmpos.mpos.data.local

import androidx.room.Database
import androidx.room.RoomDatabase

@Database(
    entities = [
        EventSnapshotEntity::class, ProductEntity::class, MemberEntity::class,
        PromotionEntity::class, PendingOrderEntity::class, ReceiptEntity::class,
    ],
    version = 9,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun eventSnapshotDao(): EventSnapshotDao
    abstract fun productDao(): ProductDao
    abstract fun memberDao(): MemberDao
    abstract fun promotionDao(): PromotionDao
    abstract fun pendingOrderDao(): PendingOrderDao
    abstract fun receiptDao(): ReceiptDao
}
