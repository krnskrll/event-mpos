package com.eventmpos.config

import com.eventmpos.event.BaEventAssignment
import com.eventmpos.event.BaEventAssignmentRepository
import com.eventmpos.event.Event
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.user.Role
import com.eventmpos.user.User
import com.eventmpos.user.UserRepository
import org.slf4j.LoggerFactory
import org.springframework.boot.ApplicationRunner
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.security.crypto.password.PasswordEncoder
import java.time.LocalDate

/**
 * Dev-only seed data so the app is usable the moment it boots. Disabled by setting SEED_ENABLED=false.
 * Idempotent: only seeds when the users table is empty.
 *
 * Seeded credentials (CHANGE before any real deployment):
 *   admin@eventmpos.local / Admin123!   (ADMIN)
 *   ba@eventmpos.local    / Ba123456!   (BA, assigned to event EV01)
 *   wh@eventmpos.local    / Wh123456!   (WAREHOUSE)
 */
@Configuration
@ConditionalOnProperty(name = ["app.seed.enabled"], havingValue = "true", matchIfMissing = false)
class DataSeeder {

    private val log = LoggerFactory.getLogger(DataSeeder::class.java)

    @Bean
    fun seedRunner(
        users: UserRepository,
        events: EventRepository,
        assignments: BaEventAssignmentRepository,
        encoder: PasswordEncoder,
        erpSyncService: com.eventmpos.product.erp.ErpSyncService,
        productService: com.eventmpos.product.ProductService,
        memberSyncService: com.eventmpos.member.sync.MemberSyncService,
        stockService: com.eventmpos.stock.StockService,
        promotionService: com.eventmpos.promotion.PromotionService,
        products: com.eventmpos.product.ProductRepository,
        variants: com.eventmpos.product.ProductVariantRepository,
    ) = ApplicationRunner {
        if (users.count() > 0L) {
            log.info("Seed skipped — users already exist")
            return@ApplicationRunner
        }
        log.warn("Seeding DEV data with default credentials — do NOT use in production")

        users.save(User(email = "admin@eventmpos.local", passwordHash = encoder.encode("Admin123!"), fullName = "System Admin", role = Role.ADMIN))
        val ba = users.save(User(email = "ba@eventmpos.local", passwordHash = encoder.encode("Ba123456!"), fullName = "Bella Advisor", role = Role.BA))
        users.save(User(email = "wh@eventmpos.local", passwordHash = encoder.encode("Wh123456!"), fullName = "Warren House", role = Role.WAREHOUSE))

        val event = events.save(
            Event(
                code = "EV01",
                name = "Mid-Year Beauty Roadshow",
                startDate = LocalDate.of(2026, 7, 1),
                endDate = LocalDate.of(2026, 7, 14),
                status = EventStatus.ACTIVE,
            ),
        )
        assignments.save(BaEventAssignment(userId = ba.id, eventId = event.id))

        // Pull the initial product catalogue from the (mock) ERP and make it all available for EV01,
        // so the BA's event snapshot has real products to browse/scan out of the box.
        val sync = erpSyncService.sync()
        productService.list().forEach { p ->
            productService.setAvailability(event.id, p.id, true)
            // allocate event stock; SKU-1004 gets a small amount to demo the low-stock warning
            val qty = if (p.sku == "SKU-1004") 8 else minOf(p.warehouseStock, 60)
            stockService.forceAllocate(event.id, p.id, qty)
        }

        // Give the lipstick a few shade variants (one with a price override) to exercise variants.
        products.findBySku("SKU-1002")?.let { lip ->
            variants.save(com.eventmpos.product.ProductVariant(product = lip, name = "Rose", sku = "SKU-1002-ROSE"))
            variants.save(com.eventmpos.product.ProductVariant(product = lip, name = "Coral", sku = "SKU-1002-CORAL"))
            variants.save(
                com.eventmpos.product.ProductVariant(
                    product = lip, name = "Plum (limited)", sku = "SKU-1002-PLUM",
                    priceOverride = java.math.BigDecimal("89000"),
                ),
            )
        }

        // Seed a few event-scoped promotions for EV01 (so the cart/quote has something to apply).
        val lipstick = productService.list().firstOrNull { it.sku == "SKU-1002" }
        promotionService.create(
            event.id,
            com.eventmpos.promotion.PromotionRequest(
                name = "10% off everything", type = com.eventmpos.promotion.PromotionType.PERCENTAGE,
                percentOff = java.math.BigDecimal("10"), priority = 1,
            ),
        )
        promotionService.create(
            event.id,
            com.eventmpos.promotion.PromotionRequest(
                name = "Rp20.000 off when you spend Rp200.000", type = com.eventmpos.promotion.PromotionType.FIXED,
                amountOff = java.math.BigDecimal("20000"), minSubtotal = java.math.BigDecimal("200000"), priority = 2,
            ),
        )
        if (lipstick != null) {
            promotionService.create(
                event.id,
                com.eventmpos.promotion.PromotionRequest(
                    name = "Lipstick bundle: buy 3, Rp30.000 off", type = com.eventmpos.promotion.PromotionType.BUNDLE,
                    amountOff = java.math.BigDecimal("30000"), scope = com.eventmpos.promotion.PromotionScope.SPECIFIC,
                    productIds = listOf(java.util.UUID.fromString(lipstick.id.toString())), minQty = 3, priority = 3,
                ),
            )
        }

        // Pull the member directory from the (mock) external member system.
        val memberSync = memberSyncService.sync()
        log.info(
            "Seed complete: 3 users, 1 event (EV01) with {} products available, {} members synced",
            sync.total, memberSync.total,
        )
    }
}
