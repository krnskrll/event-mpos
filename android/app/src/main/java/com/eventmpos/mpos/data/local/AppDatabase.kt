package com.eventmpos.mpos.data.local

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase

@Database(
    entities = [
        EventSnapshotEntity::class, ProductEntity::class, MemberEntity::class,
        PromotionEntity::class, PendingOrderEntity::class, ReceiptEntity::class,
    ],
    version = 10,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun eventSnapshotDao(): EventSnapshotDao
    abstract fun productDao(): ProductDao
    abstract fun memberDao(): MemberDao
    abstract fun promotionDao(): PromotionDao
    abstract fun pendingOrderDao(): PendingOrderDao
    abstract fun receiptDao(): ReceiptDao

    companion object {
        /** v10: promotion.priority (tie-breaker the device engine needs to match the server). */
        val MIGRATION_9_10 = object : Migration(9, 10) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL("ALTER TABLE promotion ADD COLUMN priority INTEGER NOT NULL DEFAULT 0")
            }
        }
    }
}
