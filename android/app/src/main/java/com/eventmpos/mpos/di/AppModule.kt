package com.eventmpos.mpos.di

import android.content.Context
import androidx.room.Room
import androidx.work.WorkManager
import com.eventmpos.mpos.core.scanner.ManualScannerPort
import com.eventmpos.mpos.core.scanner.ScannerPort
import com.eventmpos.mpos.core.secure.SecureTokenStorage
import com.eventmpos.mpos.data.auth.AuthRepositoryImpl
import com.eventmpos.mpos.data.event.EventRepositoryImpl
import com.eventmpos.mpos.data.local.AppDatabase
import com.eventmpos.mpos.data.local.EventSnapshotDao
import com.eventmpos.mpos.data.local.MemberDao
import com.eventmpos.mpos.data.local.PendingOrderDao
import com.eventmpos.mpos.data.local.ProductDao
import com.eventmpos.mpos.data.local.PromotionDao
import com.eventmpos.mpos.data.member.MemberRepositoryImpl
import com.eventmpos.mpos.data.order.OrderRepositoryImpl
import com.eventmpos.mpos.data.product.ProductRepositoryImpl
import com.eventmpos.mpos.data.promotion.PromotionRepositoryImpl
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.event.EventRepository
import com.eventmpos.mpos.domain.member.MemberRepository
import com.eventmpos.mpos.domain.order.OrderRepository
import com.eventmpos.mpos.domain.product.ProductRepository
import com.eventmpos.mpos.domain.promotion.PromotionRepository
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

    @Provides
    fun provideProductDao(db: AppDatabase): ProductDao = db.productDao()

    @Provides
    fun provideMemberDao(db: AppDatabase): MemberDao = db.memberDao()

    @Provides
    fun providePromotionDao(db: AppDatabase): PromotionDao = db.promotionDao()

    @Provides
    fun providePendingOrderDao(db: AppDatabase): PendingOrderDao = db.pendingOrderDao()

    @Provides
    @Singleton
    fun provideWorkManager(@ApplicationContext context: Context): WorkManager =
        WorkManager.getInstance(context)
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

    @Binds
    @Singleton
    abstract fun bindProductRepository(impl: ProductRepositoryImpl): ProductRepository

    @Binds
    @Singleton
    abstract fun bindMemberRepository(impl: MemberRepositoryImpl): MemberRepository

    @Binds
    @Singleton
    abstract fun bindPromotionRepository(impl: PromotionRepositoryImpl): PromotionRepository

    @Binds
    @Singleton
    abstract fun bindOrderRepository(impl: OrderRepositoryImpl): OrderRepository

    @Binds
    @Singleton
    abstract fun bindScannerPort(impl: ManualScannerPort): ScannerPort
}
