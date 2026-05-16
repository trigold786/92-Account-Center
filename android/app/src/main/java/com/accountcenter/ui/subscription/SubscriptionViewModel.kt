package com.accountcenter.ui.subscription

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.Subscription
import com.accountcenter.model.TierInfo
import com.accountcenter.network.ApiClient
import com.accountcenter.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SubscriptionViewModel @Inject constructor(
    private val apiClient: ApiClient,
    private val userRepository: UserRepository
) : ViewModel() {

    private val _subscriptions = MutableStateFlow<List<Subscription>>(emptyList())
    val subscriptions: StateFlow<List<Subscription>> = _subscriptions.asStateFlow()

    private val _tier = MutableStateFlow<TierInfo?>(null)
    val tier: StateFlow<TierInfo?> = _tier.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    init {
        loadData()
    }

    private fun loadData() {
        viewModelScope.launch {
            _isLoading.value = true
            val token = userRepository.getToken()
            if (token != null) {
                val userId = token.userId
                val subsResponse = apiClient.getUserSubscriptions(userId)
                if (subsResponse.isSuccessful) {
                    _subscriptions.value = subsResponse.body() ?: emptyList()
                }
                val tierResponse = apiClient.getUserTier(userId)
                if (tierResponse.isSuccessful) {
                    _tier.value = tierResponse.body()
                }
            }
            _isLoading.value = false
        }
    }

    fun refresh() {
        loadData()
    }
}
