package com.eventmpos.mpos.data.auth

import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

interface AuthApi {
    @POST("api/auth/login")
    suspend fun login(@Body body: LoginRequestDto): TokenResponseDto

    @POST("api/auth/refresh")
    suspend fun refresh(@Body body: RefreshRequestDto): TokenResponseDto

    @POST("api/auth/logout")
    suspend fun logout()

    @GET("api/me/event")
    suspend fun getMyEvent(): EventResponseDto
}
