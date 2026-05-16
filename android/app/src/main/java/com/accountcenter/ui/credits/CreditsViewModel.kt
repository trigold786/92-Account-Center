package com.accountcenter.ui.credits

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.GenerateLinkRequest
import com.accountcenter.model.ReferralSummary
import com.accountcenter.model.Transaction
import com.accountcenter.network.ApiClient
import com.accountcenter.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class CreditsViewModel @Inject constructor(
    private val apiClient: ApiClient,
    private val userRepository: UserRepository
) : ViewModel() {

    private val _balance = MutableStateFlow(0.0)
    val balance: StateFlow<Double> = _balance.asStateFlow()

    private val _transactions = MutableStateFlow<List<Transaction>>(emptyList())
    val transactions: StateFlow<List<Transaction>> = _transactions.asStateFlow()

    private val _referralSummary = MutableStateFlow<ReferralSummary?>(null)
    val referralSummary: StateFlow<ReferralSummary?> = _referralSummary.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _hasMore = MutableStateFlow(true)
    val hasMore: StateFlow<Boolean> = _hasMore.asStateFlow()

    private val _currentPage = MutableStateFlow(1)
    val currentPage: StateFlow<Int> = _currentPage.asStateFlow()

    init {
        loadData()
    }

    private fun loadData() {
        viewModelScope.launch {
            _isLoading.value = true
            val token = userRepository.getToken()
            if (token != null) {
                val userId = token.userId
                val accountResponse = apiClient.getCreditAccount(userId)
                if (accountResponse.isSuccessful) {
                    _balance.value = accountResponse.body()?.data?.balance ?: 0.0
                }
                val txResponse = apiClient.getTransactions(userId, 1)
                if (txResponse.isSuccessful) {
                    val data = txResponse.body()?.data
                    _transactions.value = data?.transactions ?: emptyList()
                    _hasMore.value = (data?.transactions?.size ?: 0) >= (data?.pageSize ?: 20)
                }
                val referralResponse = apiClient.getReferralSummary(userId)
                if (referralResponse.isSuccessful) {
                    _referralSummary.value = referralResponse.body()?.data
                }
            }
            _isLoading.value = false
        }
    }

    fun refresh() {
        _currentPage.value = 1
        loadData()
    }

    fun loadMore() {
        viewModelScope.launch {
            val token = userRepository.getToken()
            if (token != null) {
                val nextPage = _currentPage.value + 1
                val response = apiClient.getTransactions(token.userId, nextPage)
                if (response.isSuccessful) {
                    val data = response.body()?.data
                    if (data != null) {
                        _transactions.value = _transactions.value + data.transactions
                        _currentPage.value = nextPage
                        _hasMore.value = data.transactions.size >= data.pageSize
                    } else {
                        _hasMore.value = false
                    }
                }
            }
        }
    }

    fun generateReferralLink(onLink: (String) -> Unit) {
        viewModelScope.launch {
            val token = userRepository.getToken()
            if (token != null) {
                val response = apiClient.generateReferralLink(GenerateLinkRequest(token.userId.toString()))
                if (response.isSuccessful) {
                    val link = response.body()?.data?.referralLink
                    if (link != null) onLink(link)
                }
            }
        }
    }
}
