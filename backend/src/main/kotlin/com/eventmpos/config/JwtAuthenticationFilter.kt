package com.eventmpos.config

import com.eventmpos.auth.AuthPrincipal
import com.eventmpos.auth.JwtService
import com.eventmpos.user.Role
import jakarta.servlet.FilterChain
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.core.context.SecurityContextHolder
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource
import org.springframework.stereotype.Component
import org.springframework.web.filter.OncePerRequestFilter
import java.util.UUID

/**
 * Reads the Bearer access token, verifies it, and populates the SecurityContext with an
 * [AuthPrincipal] + role authority. Stateless: no session is created.
 */
@Component
class JwtAuthenticationFilter(
    private val jwtService: JwtService,
) : OncePerRequestFilter() {

    override fun doFilterInternal(
        request: HttpServletRequest,
        response: HttpServletResponse,
        filterChain: FilterChain,
    ) {
        val header = request.getHeader("Authorization")
        if (header != null && header.startsWith("Bearer ")) {
            val claims = jwtService.parse(header.substring(7))
            if (claims != null && SecurityContextHolder.getContext().authentication == null) {
                val principal = AuthPrincipal(
                    userId = UUID.fromString(claims.subject),
                    email = claims["email"] as String,
                    fullName = claims["name"] as? String ?: "",
                    role = Role.valueOf(claims["role"] as String),
                )
                val authority = SimpleGrantedAuthority("ROLE_${principal.role.name}")
                val auth = UsernamePasswordAuthenticationToken(principal, null, listOf(authority))
                auth.details = WebAuthenticationDetailsSource().buildDetails(request)
                SecurityContextHolder.getContext().authentication = auth
            }
        }
        filterChain.doFilter(request, response)
    }
}
