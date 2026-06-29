package com.eventmpos.user

/**
 * The three system roles. Access boundaries are enforced per-role in [com.eventmpos.config.SecurityConfig].
 * Spring Security authorities are prefixed with "ROLE_" (e.g. ROLE_ADMIN).
 */
enum class Role {
    ADMIN,
    BA,
    WAREHOUSE,
}
