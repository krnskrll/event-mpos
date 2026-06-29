package com.eventmpos.mpos.data.local

import androidx.room.Database
import androidx.room.RoomDatabase

@Database(
    entities = [EventSnapshotEntity::class, ProductEntity::class],
    version = 2,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun eventSnapshotDao(): EventSnapshotDao
    abstract fun productDao(): ProductDao
}
