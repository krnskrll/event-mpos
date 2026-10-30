package com.eventmpos.mpos.ui.receipt

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.core.printer.PrintResult
import com.eventmpos.mpos.domain.receipt.ReceiptFormatter
import com.eventmpos.mpos.domain.receipt.ReceiptRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ReceiptUiState(
    val lines: List<String> = emptyList(),
    val loading: Boolean = true,
    val printing: Boolean = false,
    val message: String? = null,
    val error: Boolean = false,
) {
    val cancelled: Boolean get() = lines.firstOrNull() == ReceiptFormatter.CANCELLED
}

@HiltViewModel
class ReceiptViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val receiptRepository: ReceiptRepository,
) : ViewModel() {

    private val orderKey: String = checkNotNull(savedStateHandle["orderKey"])
    private val _state = MutableStateFlow(ReceiptUiState())
    val state: StateFlow<ReceiptUiState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            val r = receiptRepository.get(orderKey)
            _state.update {
                it.copy(
                    loading = false,
                    lines = r?.lines ?: emptyList(),
                    message = if (r == null) "No stored receipt for this order." else null,
                    error = r == null,
                )
            }
        }
    }

    /** Reprint the stored receipt — no new order, no network. */
    fun reprint() {
        if (_state.value.printing) return
        _state.update { it.copy(printing = true, message = null, error = false) }
        viewModelScope.launch {
            when (val res = receiptRepository.reprint(orderKey)) {
                is PrintResult.Success -> _state.update { it.copy(printing = false, message = "Printed ✓", error = false) }
                is PrintResult.Failure -> _state.update { it.copy(printing = false, message = "Print failed: ${res.message}", error = true) }
            }
        }
    }
}
