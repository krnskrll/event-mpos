package com.eventmpos.user

import com.eventmpos.common.ApiException
import org.springframework.http.HttpStatus
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.util.UUID

@Service
class UserService(
    private val userRepository: UserRepository,
    private val passwordEncoder: PasswordEncoder,
) {

    fun listUsers(): List<UserResponse> = userRepository.findAll().map { it.toResponse() }

    fun listByRole(role: Role): List<UserResponse> =
        userRepository.findAll().filter { it.role == role }.map { it.toResponse() }

    @Transactional
    fun createUser(request: CreateUserRequest): UserResponse {
        val email = request.email.trim().lowercase()
        if (userRepository.existsByEmail(email)) {
            throw ApiException(HttpStatus.CONFLICT, "A user with email '$email' already exists")
        }
        val user = User(
            email = email,
            passwordHash = passwordEncoder.encode(request.password),
            fullName = request.fullName.trim(),
            role = request.role,
        )
        return userRepository.save(user).toResponse()
    }

    @Transactional
    fun setActive(userId: UUID, active: Boolean): UserResponse {
        val user = userRepository.findById(userId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "User not found")
        }
        user.active = active
        return userRepository.save(user).toResponse()
    }

    /** Partial update of a user's name / role / active flag (admin-only). */
    @Transactional
    fun updateUser(userId: UUID, request: UpdateUserRequest): UserResponse {
        val user = userRepository.findById(userId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "User not found")
        }
        request.fullName?.let { user.fullName = it.trim() }
        request.role?.let { user.role = it }
        request.active?.let { user.active = it }
        return userRepository.save(user).toResponse()
    }
}
