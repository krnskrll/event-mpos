package com.eventmpos.mpos.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.eventmpos.mpos.data.cart.CartStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import javax.inject.Inject

/** Exposes the cart item count for the bottom-nav badge. */
@HiltViewModel
class CartBadgeViewModel @Inject constructor(cartStore: CartStore) : ViewModel() {
    val count: StateFlow<Int> =
        cartStore.items.map { it.values.sum() }
            .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)
}
