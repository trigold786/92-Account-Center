package com.accountcenter.ui.security

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.PushDevice
import com.accountcenter.model.RiskEvent
import com.accountcenter.network.ApiClient
import com.accountcenter.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import androidx.compose.runtime.mutableStateOf
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SecurityViewModel @Inject constructor(
    private val apiClient: ApiClient,
    private val userRepository: UserRepository
) : ViewModel() {

    private val _riskEvents = MutableStateFlow<List<RiskEvent>>(emptyList())
    val riskEvents: StateFlow<List<RiskEvent>> = _riskEvents.asStateFlow()

    private val _devices = MutableStateFlow<List<PushDevice>>(emptyList())
    val devices: StateFlow<List<PushDevice>> = _devices.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    var currentPassword = mutableStateOf("")
    var newPassword = mutableStateOf("")
    var confirmPassword = mutableStateOf("")
    var passwordChangeMessage = mutableStateOf<String?>(null)
    var passwordChangeSuccess = mutableStateOf(false)

    init {
        loadData()
    }

    fun changePassword() {
        passwordChangeMessage.value = "密码修改功能暂未开放"
        passwordChangeSuccess.value = false
    }

    private fun loadData() {
        viewModelScope.launch {
            _isLoading.value = true
            val token = userRepository.getToken()
            if (token != null) {
                val userId = token.userId
                val riskResponse = apiClient.getRiskHistory(userId)
                if (riskResponse.isSuccessful) {
                    _riskEvents.value = riskResponse.body()?.data?.events ?: emptyList()
                }
                val deviceResponse = apiClient.getUserDevices(userId)
                if (deviceResponse.isSuccessful) {
                    _devices.value = deviceResponse.body()?.data?.devices ?: emptyList()
                }
            }
            _isLoading.value = false
        }
    }

}
