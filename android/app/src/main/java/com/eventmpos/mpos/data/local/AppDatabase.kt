package com.eventmpos.mpos.data.local

import androidx.room.Database
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase

@Database(
    entities = [
        EventSnapshotEntity::class, ProductEntity::class, MemberEntity::class,
        PromotionEntity::class, PendingOrderEntity::class, ReceiptEntity::class, CartDraftEntity::class,
    ],
    version = 11,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun eventSnapshotDao(): EventSnapshotDao
    abstract fun productDao(): ProductDao
    abstract fun memberDao(): MemberDao
    abstract fun promotionDao(): PromotionDao
    abstract fun pendingOrderDao(): PendingOrderDao
    abstract fun receiptDao(): ReceiptDao
    abstract fun cartDraftDao(): CartDraftDao

    companion object {
        /** v10: promotion.priority (tie-breaker the device engine needs to match the server). */
        val MIGRATION_9_10 = object : Migration(9, 10) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL("ALTER TABLE promotion ADD COLUMN priority INTEGER NOT NULL DEFAULT 0")
            }
        }

        /** v11: cart_draft (saved drafts + the autosaved live cart). */
        val MIGRATION_10_11 = object : Migration(10, 11) {
            override fun migrate(db: SupportSQLiteDatabase) {
                db.execSQL(
                    "CREATE TABLE IF NOT EXISTS `cart_draft` (`id` TEXT NOT NULL, `eventId` TEXT NOT NULL, " +
                        "`label` TEXT NOT NULL, `customerJson` TEXT, `linesJson` TEXT NOT NULL, " +
                        "`createdAtEpochMs` INTEGER NOT NULL, `live` INTEGER NOT NULL, PRIMARY KEY(`id`))",
                )
                db.execSQL("CREATE INDEX IF NOT EXISTS `index_cart_draft_eventId` ON `cart_draft` (`eventId`)")
            }
        }
    }
}
