package com.eventmpos.common

import org.springframework.http.HttpStatus

/** Application-level error carrying the HTTP status to return. Handled by [GlobalExceptionHandler]. */
class ApiException(
    val status: HttpStatus,
    override val message: String,
) : RuntimeException(message)
