package com.eventmpos.member.sync

import com.eventmpos.member.MemberTier
import org.springframework.beans.factory.annotation.Value
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.stereotype.Component
import org.springframework.web.client.RestClient

/**
 * REST-pull member source. Calls a configured endpoint (`app.member.rest-url`) returning member
 * JSON and maps it into [MemberRecord]. Adjust [MemberJson] to match the real member system.
 */
@Component
@ConditionalOnProperty(name = ["app.member.source"], havingValue = "rest")
class RestMemberSource(
    @Value("\${app.member.rest-url:}") private val restUrl: String,
) : MemberSource {

    private val client = RestClient.create()

    override val description get() = "rest:$restUrl"

    override fun fetchMembers(): List<MemberRecord> {
        val body = client.get().uri(restUrl).retrieve().body(Array<MemberJson>::class.java) ?: emptyArray()
        return body.map {
            MemberRecord(
                memberCode = it.memberCode,
                qrCode = it.qrCode,
                name = it.name,
                tier = runCatching { MemberTier.valueOf(it.tier.uppercase()) }.getOrDefault(MemberTier.STANDARD),
                region = it.region,
                active = it.active,
            )
        }
    }

    data class MemberJson(
        val memberCode: String = "",
        val qrCode: String = "",
        val name: String = "",
        val tier: String = "STANDARD",
        val region: String? = null,
        val active: Boolean = true,
    )
}
