package com.accountcenter.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.LoginRequest
import com.accountcenter.model.SMSSendRequest
import com.accountcenter.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val phoneNumber: String = "",
    val password: String = "",
    val verificationCode: String = "",
    val loginMode: LoginMode = LoginMode.PASSWORD,
    val isLoading: Boolean = false,
    val countdownSeconds: Int = 0,
    val errorMessage: String? = null
)

enum class LoginMode {
    PASSWORD, VERIFICATION_CODE
}

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {
    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    private var countdownJob: kotlinx.coroutines.Job? = null

    fun onPhoneNumberChange(value: String) {
        _uiState.value = _uiState.value.copy(phoneNumber = value, errorMessage = null)
    }

    fun onPasswordChange(value: String) {
        _uiState.value = _uiState.value.copy(password = value, errorMessage = null)
    }

    fun onVerificationCodeChange(value: String) {
        _uiState.value = _uiState.value.copy(verificationCode = value, errorMessage = null)
    }

    fun onLoginModeChange(mode: LoginMode) {
        _uiState.value = _uiState.value.copy(loginMode = mode)
    }

    fun login(onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val state = _uiState.value
            val request = if (state.loginMode == LoginMode.PASSWORD) {
                LoginRequest.withPassword(state.phoneNumber, state.password)
            } else {
                LoginRequest.withCode(state.phoneNumber, state.verificationCode)
            }
            val result = authRepository.login(request)

            if (result.isSuccess) {
                onSuccess()
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.exceptionOrNull()?.message ?: "登录失败"
                )
            }
        }
    }

    fun sendVerificationCode() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val request = SMSSendRequest(_uiState.value.phoneNumber)
            val result = authRepository.sendSMS(request)

            if (result.isSuccess) {
                startCountdown()
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.exceptionOrNull()?.message ?: "发送失败"
                )
            }
        }
    }

    private fun startCountdown() {
        _uiState.value = _uiState.value.copy(isLoading = false, countdownSeconds = 60)
        countdownJob?.cancel()
        countdownJob = viewModelScope.launch {
            while (_uiState.value.countdownSeconds > 0) {
                delay(1000)
                _uiState.value = _uiState.value.copy(countdownSeconds = _uiState.value.countdownSeconds - 1)
            }
        }
    }
}
