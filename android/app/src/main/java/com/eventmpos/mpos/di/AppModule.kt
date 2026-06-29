package com.eventmpos.mpos.di

import android.content.Context
import androidx.room.Room
import com.eventmpos.mpos.core.secure.SecureTokenStorage
import com.eventmpos.mpos.data.auth.AuthRepositoryImpl
import com.eventmpos.mpos.data.event.EventRepositoryImpl
import com.eventmpos.mpos.data.local.AppDatabase
import com.eventmpos.mpos.data.local.EventSnapshotDao
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.event.EventRepository
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppProvidesModule {

    @Provides
    @Singleton
    fun provideSecureTokenStorage(@ApplicationContext context: Context): SecureTokenStorage =
        SecureTokenStorage(context)

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): AppDatabase =
        Room.databaseBuilder(context, AppDatabase::class.java, "eventmpos.db")
            .fallbackToDestructiveMigration()
            .build()

    @Provides
    fun provideEventSnapshotDao(db: AppDatabase): EventSnapshotDao = db.eventSnapshotDao()
}

@Module
@InstallIn(SingletonComponent::class)
abstract class AppBindsModule {

    @Binds
    @Singleton
    abstract fun bindAuthRepository(impl: AuthRepositoryImpl): AuthRepository

    @Binds
    @Singleton
    abstract fun bindEventRepository(impl: EventRepositoryImpl): EventRepository
}
