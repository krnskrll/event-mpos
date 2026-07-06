package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.data.order.OrderConfirmationDto
import com.eventmpos.mpos.data.order.OrderSummaryDto
import com.eventmpos.mpos.data.order.SubmitOrderRequestDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

interface AuthApi {
    @POST("api/auth/login")
    suspend fun login(@Body body: LoginRequestDto): TokenResponseDto

    @POST("api/auth/refresh")
    suspend fun refresh(@Body body: RefreshRequestDto): TokenResponseDto

    @POST("api/auth/logout")
    suspend fun logout()

    @GET("api/me/event")
    suspend fun getMyEvent(): EventResponseDto

    @GET("api/me/events")
    suspend fun getMyEvents(): List<EventResponseDto>

    @GET("api/me/events/{eventId}/snapshot")
    suspend fun getSnapshot(
        @Path("eventId") eventId: String,
        @Query("version") version: String?,
    ): SnapshotEnvelopeDto

    @GET("api/me/events/{eventId}/snapshot/hash")
    suspend fun getSnapshotHash(@Path("eventId") eventId: String): SnapshotHashDto

    @GET("api/me/members/by-qr/{qrCode}")
    suspend fun getMemberByQr(@Path("qrCode") qrCode: String): MemberLookupDto

    @GET("api/me/events/{eventId}/stock")
    suspend fun getStockLevels(@Path("eventId") eventId: String): List<StockLevelDto>

    @POST("api/me/events/{eventId}/orders")
    suspend fun submitOrder(
        @Path("eventId") eventId: String,
        @Body body: SubmitOrderRequestDto,
    ): OrderConfirmationDto

    @GET("api/me/events/{eventId}/orders")
    suspend fun getOrderHistory(@Path("eventId") eventId: String): List<OrderSummaryDto>

    /** Fire-and-forget: log that a receipt was printed/reprinted for an order (backend audit). */
    @POST("api/me/orders/{orderId}/receipts")
    suspend fun logReceipt(@Path("orderId") orderId: String)
}
