package com.eventmpos.mpos.ui.promotions

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.domain.promotion.Promotion
import com.eventmpos.mpos.domain.promotion.PromotionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

@HiltViewModel
class PromotionsViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    promotionRepository: PromotionRepository,
) : ViewModel() {

    private val eventId: String = checkNotNull(savedStateHandle["eventId"])

    val promotions: StateFlow<List<Promotion>> =
        promotionRepository.observePromotions(eventId)
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())
}
