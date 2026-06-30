package com.eventmpos.config

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.http.HttpMethod
import org.springframework.security.config.annotation.web.builders.HttpSecurity
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer
import org.springframework.security.config.http.SessionCreationPolicy
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.security.web.SecurityFilterChain
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter
import org.springframework.web.cors.CorsConfiguration
import org.springframework.web.cors.CorsConfigurationSource
import org.springframework.web.cors.UrlBasedCorsConfigurationSource

/**
 * Central place where role-based access boundaries are declared. Adding a new protected endpoint
 * later is just one more matcher here (or a @PreAuthorize on the controller) — the JWT filter and
 * stateless setup stay untouched. This keeps EPIC-01's promise: "structured so adding new protected
 * endpoints later is straightforward".
 */
@Configuration
class SecurityConfig(
    private val jwtAuthenticationFilter: JwtAuthenticationFilter,
) {

    @Bean
    fun passwordEncoder(): PasswordEncoder = BCryptPasswordEncoder()

    @Bean
    fun securityFilterChain(http: HttpSecurity): SecurityFilterChain {
        http
            .csrf { it.disable() }
            .cors { it.configurationSource(corsConfigurationSource()) }
            .sessionManagement { it.sessionCreationPolicy(SessionCreationPolicy.STATELESS) }
            .authorizeHttpRequests { auth ->
                auth
                    // Public: login + refresh + health + the WebSocket handshake (STOMP CONNECT
                    // carries order data only within the warehouse context)
                    .requestMatchers("/api/auth/login", "/api/auth/refresh").permitAll()
                    .requestMatchers("/ws/**", "/ws-native/**").permitAll()
                    .requestMatchers(HttpMethod.GET, "/actuator/health", "/api/health").permitAll()
                    // Admin-only: user management + BA assignment
                    .requestMatchers("/api/admin/**").hasRole("ADMIN")
                    // Warehouse area (fleshed out in EPIC-09)
                    .requestMatchers("/api/warehouse/**").hasAnyRole("WAREHOUSE", "ADMIN")
                    // Everything else just needs a valid token
                    .anyRequest().authenticated()
            }
            .exceptionHandling { ex ->
                // Missing/invalid token -> 401.
                ex.authenticationEntryPoint { _, response, _ ->
                    response.status = 401
                    response.contentType = "application/json"
                    response.writer.write("""{"status":401,"error":"Unauthorized","message":"Authentication required"}""")
                }
                // Authenticated but lacking the required role -> 403.
                ex.accessDeniedHandler { _, response, _ ->
                    response.status = 403
                    response.contentType = "application/json"
                    response.writer.write("""{"status":403,"error":"Forbidden","message":"You do not have access to this resource"}""")
                }
            }
            .addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter::class.java)
            .httpBasic { it.disable() }
            .formLogin { it.disable() }
        return http.build()
    }

    @Bean
    fun corsConfigurationSource(): CorsConfigurationSource {
        val config = CorsConfiguration().apply {
            // Web dashboard dev origin; tighten/extend for production deployments.
            allowedOriginPatterns = listOf("http://localhost:*", "http://127.0.0.1:*")
            allowedMethods = listOf("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS")
            allowedHeaders = listOf("*")
            allowCredentials = true
        }
        return UrlBasedCorsConfigurationSource().apply { registerCorsConfiguration("/**", config) }
    }
}
