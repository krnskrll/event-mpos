package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.core.secure.SecureTokenStorage
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject

/**
 * Attaches the stored access token to outgoing requests, except the auth endpoints themselves
 * (login/refresh are public and must not carry a possibly-expired token).
 */
class AuthInterceptor @Inject constructor(
    private val tokenStorage: SecureTokenStorage,
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val path = request.url.encodedPath
        if (path.contains("/api/auth/login") || path.contains("/api/auth/refresh")) {
            return chain.proceed(request)
        }
        val token = tokenStorage.accessToken
        val authed = if (token != null) {
            request.newBuilder().header("Authorization", "Bearer $token").build()
        } else {
            request
        }
        return chain.proceed(authed)
    }
}
