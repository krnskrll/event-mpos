package com.eventmpos.config

import com.eventmpos.event.BaEventAssignment
import com.eventmpos.event.BaEventAssignmentRepository
import com.eventmpos.event.Event
import com.eventmpos.event.EventRepository
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
            ),
        )
        assignments.save(BaEventAssignment(userId = ba.id, eventId = event.id))
        log.info("Seed complete: 3 users, 1 event (EV01), BA assigned to EV01")
    }
}
