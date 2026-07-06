package com.eventmpos.mpos.ui.customer

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.cart.CartStore
import com.eventmpos.mpos.domain.member.CustomerSource
import com.eventmpos.mpos.domain.member.MemberRepository
import com.eventmpos.mpos.domain.member.SelectedCustomer
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class CustomerUiState(
    val identifying: Boolean = false,
    val selected: SelectedCustomer? = null,
    /** Set when a QR scan/typed code didn't resolve anywhere — BA falls back to manual entry. */
    val notFoundQr: String? = null,
)

@HiltViewModel
class CustomerViewModel @Inject constructor(
    private val memberRepository: MemberRepository,
    private val cartStore: CartStore,
) : ViewModel() {

    private val _state = MutableStateFlow(CustomerUiState())
    val state: StateFlow<CustomerUiState> = _state.asStateFlow()

    /** Resolve a scanned/typed QR: local cache → online fallback → not-found (manual). */
    fun identifyByQr(qrCode: String) {
        val code = qrCode.trim()
        if (code.isEmpty()) return
        _state.update { it.copy(identifying = true, notFoundQr = null) }
        viewModelScope.launch {
            val member = memberRepository.identify(code)
            if (member != null) {
                val source =
                    if (memberRepository.cachedCount() > 0) CustomerSource.LOCAL_CACHE else CustomerSource.ONLINE_LOOKUP
                val selected = SelectedCustomer(member.name, member.tier, member.qrCode, source)
                cartStore.setCustomer(selected)
                _state.update { it.copy(identifying = false, selected = selected, notFoundQr = null) }
            } else {
                _state.update { it.copy(identifying = false, notFoundQr = code) }
            }
        }
    }

    /** Manual fallback — BA types the customer's name to proceed when QR can't be resolved. */
    fun setManual(name: String) {
        if (name.isBlank()) return
        val selected = SelectedCustomer(name.trim(), null, null, CustomerSource.MANUAL)
        cartStore.setCustomer(selected)
        _state.update { it.copy(selected = selected, notFoundQr = null) }
    }

    fun clear() {
        cartStore.setCustomer(null)
        _state.update { CustomerUiState() }
    }
}
