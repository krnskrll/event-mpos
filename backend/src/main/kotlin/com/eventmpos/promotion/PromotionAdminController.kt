package com.eventmpos.promotion

import jakarta.validation.Valid
import org.springframework.http.HttpStatus
import org.springframework.web.bind.annotation.DeleteMapping
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.PutMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.ResponseStatus
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

/** Admin: manage event-scoped promotions (under /api/admin, ROLE_ADMIN). */
@RestController
@RequestMapping("/api/admin")
class PromotionAdminController(
    private val promotionService: PromotionService,
) {
    @GetMapping("/events/{eventId}/promotions")
    fun list(@PathVariable eventId: UUID): List<PromotionResponse> =
        promotionService.listForEvent(eventId)

    @PostMapping("/events/{eventId}/promotions")
    @ResponseStatus(HttpStatus.CREATED)
    fun create(@PathVariable eventId: UUID, @Valid @RequestBody req: PromotionRequest): PromotionResponse =
        promotionService.create(eventId, req)

    @PutMapping("/promotions/{id}")
    fun update(@PathVariable id: UUID, @Valid @RequestBody req: PromotionRequest): PromotionResponse =
        promotionService.update(id, req)

    @DeleteMapping("/promotions/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    fun delete(@PathVariable id: UUID) = promotionService.delete(id)
}
