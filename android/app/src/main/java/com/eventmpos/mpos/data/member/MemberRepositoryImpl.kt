package com.eventmpos.mpos.data.member

import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.local.MemberDao
import com.eventmpos.mpos.domain.member.Member
import com.eventmpos.mpos.domain.member.MemberRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class MemberRepositoryImpl @Inject constructor(
    private val dao: MemberDao,
    private val api: AuthApi,
) : MemberRepository {

    override suspend fun identify(qrCode: String): Member? = withContext(Dispatchers.IO) {
        val code = qrCode.trim()
        // 1) local cache — instant, works offline
        dao.findByQr(code)?.let { return@withContext Member(it.qrCode, it.name, it.tier) }
        // 2) online fallback — only if we have network; any failure -> treat as not found
        runCatching {
            val dto = api.getMemberByQr(code)
            Member(dto.qrCode, dto.name, dto.tier)
        }.getOrNull()
    }

    override suspend fun cachedCount(): Int = withContext(Dispatchers.IO) { dao.count() }
}
