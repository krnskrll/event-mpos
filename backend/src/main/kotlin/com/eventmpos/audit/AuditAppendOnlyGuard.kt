package com.eventmpos.audit

import org.slf4j.LoggerFactory
import org.springframework.boot.context.event.ApplicationReadyEvent
import org.springframework.context.event.EventListener
import org.springframework.jdbc.core.JdbcTemplate
import org.springframework.stereotype.Component

/**
 * Enforces the append-only rule on `order_audit_log` at the database level, in addition to the
 * code-level immutability (the entity has no mutators). A PostgreSQL BEFORE UPDATE/DELETE trigger
 * raises an exception, so the history can never be altered even by a direct SQL statement.
 *
 * Installed once on startup. On non-PostgreSQL databases (e.g. H2 in tests) plpgsql isn't available,
 * so the attempt is caught and skipped — the code-level guard still applies there.
 */
@Component
class AuditAppendOnlyGuard(private val jdbcTemplate: JdbcTemplate) {

    private val log = LoggerFactory.getLogger(AuditAppendOnlyGuard::class.java)

    @EventListener(ApplicationReadyEvent::class)
    fun install() {
        runCatching {
            jdbcTemplate.execute(
                """
                CREATE OR REPLACE FUNCTION reject_order_audit_mutation() RETURNS trigger AS ${'$'}${'$'}
                BEGIN
                    RAISE EXCEPTION 'order_audit_log is append-only';
                END;
                ${'$'}${'$'} LANGUAGE plpgsql;
                """.trimIndent(),
            )
            jdbcTemplate.execute("DROP TRIGGER IF EXISTS trg_order_audit_append_only ON order_audit_log;")
            jdbcTemplate.execute(
                """
                CREATE TRIGGER trg_order_audit_append_only
                    BEFORE UPDATE OR DELETE ON order_audit_log
                    FOR EACH ROW EXECUTE FUNCTION reject_order_audit_mutation();
                """.trimIndent(),
            )
            log.info("Installed append-only trigger on order_audit_log")
        }.onFailure {
            log.warn("Append-only DB trigger not installed ({}); code-level immutability still applies", it.message)
        }
    }
}
