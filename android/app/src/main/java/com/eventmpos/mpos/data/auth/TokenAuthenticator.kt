package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.core.secure.SecureTokenStorage
import kotlinx.serialization.json.Json
import okhttp3.Authenticator
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import okhttp3.Route
import javax.inject.Inject
import javax.inject.Named
import javax.inject.Singleton

/**
 * Implements "token is refreshed automatically when it expires" (EPIC-01).
 *
 * OkHttp calls [authenticate] whenever a request comes back 401. We exchange the stored refresh
 * token for a fresh pair (rotation happens server-side), persist them, and retry the original
 * request with the new access token. If refresh fails, we clear the session and give up (returning
 * null), which surfaces the 401 to the caller and routes the user back to login.
 *
 * Uses a bare OkHttpClient (no AuthInterceptor / no Authenticator) for the refresh call so it can
 * never recurse. Guarded by a lock + a check that the token actually changed, so a burst of
 * concurrent 401s triggers exactly one refresh.
 */
@Singleton
class TokenAuthenticator @Inject constructor(
    private val tokenStorage: SecureTokenStorage,
    @Named("baseUrl") private val baseUrl: String,
    private val json: Json,
) : Authenticator {

    private val bareClient = OkHttpClient()
    private val lock = Any()

    override fun authenticate(route: Route?, response: Response): Request? {
        val failedToken = response.request.header("Authorization")?.removePrefix("Bearer ")

        synchronized(lock) {
            val current = tokenStorage.accessToken
            // Another thread already refreshed since this request was sent — just retry with it.
            if (current != null && current != failedToken) {
                return response.request.newBuilder()
                    .header("Authorization", "Bearer $current")
                    .build()
            }

            val refresh = tokenStorage.refreshToken ?: return null
            val newTokens = runCatching { refreshSync(refresh) }.getOrNull()
            if (newTokens == null) {
                tokenStorage.clear()
                return null
            }
            tokenStorage.saveTokens(newTokens.accessToken, newTokens.refreshToken)
            return response.request.newBuilder()
                .header("Authorization", "Bearer ${newTokens.accessToken}")
                .build()
        }
    }

    private fun refreshSync(refreshToken: String): TokenResponseDto? {
        val payload = json.encodeToString(RefreshRequestDto.serializer(), RefreshRequestDto(refreshToken))
        val request = Request.Builder()
            .url(baseUrl.trimEnd('/') + "/api/auth/refresh")
            .post(payload.toRequestBody("application/json".toMediaType()))
            .build()
        bareClient.newCall(request).execute().use { resp ->
            if (!resp.isSuccessful) return null
            val body = resp.body?.string() ?: return null
            return json.decodeFromString(TokenResponseDto.serializer(), body)
        }
    }
}
