package com.eventmpos.mpos.data.auth

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
}
