package com.accountcenter.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.RFMScore
import com.accountcenter.model.UserDisplay
import com.accountcenter.network.ApiClient
import com.accountcenter.repository.AuthRepository
import com.accountcenter.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val apiClient: ApiClient,
    userRepository: UserRepository
) : ViewModel() {
    val userDisplay: StateFlow<UserDisplay> = userRepository.currentUser.map { user ->
        UserDisplay(
            accountId = user?.accountId ?: "",
            phoneNumber = user?.phoneNumber ?: ""
        )
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), UserDisplay())

    private val _rfmScore = MutableStateFlow<RFMScore?>(null)
    val rfmScore: StateFlow<RFMScore?> = _rfmScore.asStateFlow()

    init {
        viewModelScope.launch {
            val token = userRepository.getToken()
            if (token != null) {
                val response = apiClient.getRFMScore(token.userId)
                if (response.isSuccessful) {
                    _rfmScore.value = response.body()
                }
            }
        }
    }

    fun logout(onLogout: () -> Unit) {
        viewModelScope.launch {
            authRepository.logout()
            onLogout()
        }
    }
}
