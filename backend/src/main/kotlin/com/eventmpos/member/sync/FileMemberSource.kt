package com.eventmpos.member.sync

import com.eventmpos.common.ApiException
import com.eventmpos.member.MemberTier
import org.springframework.beans.factory.annotation.Value
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.core.io.DefaultResourceLoader
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Component

/**
 * File-import member source. Reads a CSV export.
 * Columns: memberCode,qrCode,name,tier,region,active   (configure via `app.member.file-path`).
 */
@Component
@ConditionalOnProperty(name = ["app.member.source"], havingValue = "file")
class FileMemberSource(
    @Value("\${app.member.file-path:classpath:members/members.csv}")
    private val filePath: String,
) : MemberSource {

    override val description get() = "file:$filePath"

    override fun fetchMembers(): List<MemberRecord> {
        val resource = DefaultResourceLoader().getResource(filePath)
        if (!resource.exists()) throw ApiException(HttpStatus.BAD_REQUEST, "Member file not found at $filePath")
        val lines = resource.inputStream.bufferedReader().readLines().filter { it.isNotBlank() }
        if (lines.isEmpty()) return emptyList()
        return lines.drop(1).map { line ->
            val c = line.split(",").map { it.trim() }
            require(c.size >= 6) { "Malformed member CSV row (need 6 columns): $line" }
            MemberRecord(
                memberCode = c[0],
                qrCode = c[1],
                name = c[2],
                tier = runCatching { MemberTier.valueOf(c[3].uppercase()) }.getOrDefault(MemberTier.STANDARD),
                region = c[4].ifBlank { null },
                active = c[5].equals("true", ignoreCase = true),
            )
        }
    }
}
