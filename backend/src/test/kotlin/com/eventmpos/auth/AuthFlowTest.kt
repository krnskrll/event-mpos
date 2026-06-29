package com.eventmpos.auth

import com.eventmpos.user.CreateUserRequest
import com.eventmpos.user.Role
import com.eventmpos.user.UserService
import com.fasterxml.jackson.databind.ObjectMapper
import org.junit.jupiter.api.Assertions.assertNotEquals
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.MediaType
import org.springframework.test.context.TestPropertySource
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath
import org.springframework.test.web.servlet.result.MockMvcResultMatchers.status

/**
 * End-to-end auth flow against an in-memory H2 DB: login -> access protected route ->
 * refresh (with rotation) -> role boundaries. Run with: gradle test
 */
@SpringBootTest
@AutoConfigureMockMvc
@TestPropertySource(
    properties = [
        "spring.datasource.url=jdbc:h2:mem:authtest;MODE=PostgreSQL;DB_CLOSE_DELAY=-1",
        "spring.datasource.driver-class-name=org.h2.Driver",
        "spring.datasource.username=sa",
        "spring.datasource.password=",
        "spring.jpa.hibernate.ddl-auto=create-drop",
        "app.seed.enabled=false",
    ],
)
class AuthFlowTest(
    @Autowired val mvc: MockMvc,
    @Autowired val userService: UserService,
    @Autowired val mapper: ObjectMapper,
) {

    @BeforeEach
    fun seed() {
        if (!emailExists("admin@test.local")) {
            userService.createUser(CreateUserRequest("admin@test.local", "Admin123!", "Admin", Role.ADMIN))
        }
        if (!emailExists("ba@test.local")) {
            userService.createUser(CreateUserRequest("ba@test.local", "Ba123456!", "BA", Role.BA))
        }
    }

    private fun emailExists(email: String) =
        runCatching { userService.listUsers().any { it.email == email } }.getOrDefault(false)

    @Test
    fun `login returns tokens and protects routes`() {
        val tokens = login("admin@test.local", "Admin123!")

        // /api/me requires a valid token
        mvc.perform(get("/api/me").header("Authorization", "Bearer ${tokens.accessToken}"))
            .andExpect(status().isOk)
            .andExpect(jsonPath("$.role").value("ADMIN"))

        // No token -> 401
        mvc.perform(get("/api/me")).andExpect(status().isUnauthorized)
    }

    @Test
    fun `wrong password is rejected`() {
        mvc.perform(
            post("/api/auth/login").contentType(MediaType.APPLICATION_JSON)
                .content("""{"email":"admin@test.local","password":"nope"}"""),
        ).andExpect(status().isUnauthorized)
    }

    @Test
    fun `refresh rotates the refresh token`() {
        val tokens = login("admin@test.local", "Admin123!")
        val refreshed = mapper.readValue(
            mvc.perform(
                post("/api/auth/refresh").contentType(MediaType.APPLICATION_JSON)
                    .content("""{"refreshToken":"${tokens.refreshToken}"}"""),
            ).andExpect(status().isOk).andReturn().response.contentAsString,
            TokenResponse::class.java,
        )
        assertNotEquals(tokens.refreshToken, refreshed.refreshToken)

        // Old refresh token is now revoked
        mvc.perform(
            post("/api/auth/refresh").contentType(MediaType.APPLICATION_JSON)
                .content("""{"refreshToken":"${tokens.refreshToken}"}"""),
        ).andExpect(status().isUnauthorized)
    }

    @Test
    fun `BA cannot reach admin routes`() {
        val tokens = login("ba@test.local", "Ba123456!")
        mvc.perform(get("/api/admin/users").header("Authorization", "Bearer ${tokens.accessToken}"))
            .andExpect(status().isForbidden)
    }

    private fun login(email: String, password: String): TokenResponse =
        mapper.readValue(
            mvc.perform(
                post("/api/auth/login").contentType(MediaType.APPLICATION_JSON)
                    .content("""{"email":"$email","password":"$password"}"""),
            ).andExpect(status().isOk).andReturn().response.contentAsString,
            TokenResponse::class.java,
        )
}
