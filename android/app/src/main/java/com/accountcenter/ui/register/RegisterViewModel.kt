package com.accountcenter.ui.register

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.model.LoginRequest
import com.accountcenter.model.RegisterRequest
import com.accountcenter.model.SMSSendRequest
import com.accountcenter.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class RegisterUiState(
    val phoneNumber: String = "",
    val verificationCode: String = "",
    val password: String = "",
    val confirmPassword: String = "",
    val accountId: String = "",
    val agreeToTerms: Boolean = false,
    val referralCode: String = "",
    val isLoading: Boolean = false,
    val countdownSeconds: Int = 0,
    val errorMessage: String? = null
)

@HiltViewModel
class RegisterViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {
    private val _uiState = MutableStateFlow(RegisterUiState())
    val uiState: StateFlow<RegisterUiState> = _uiState.asStateFlow()

    private var countdownJob: kotlinx.coroutines.Job? = null

    fun onPhoneNumberChange(value: String) {
        _uiState.value = _uiState.value.copy(phoneNumber = value, errorMessage = null)
    }

    fun onVerificationCodeChange(value: String) {
        _uiState.value = _uiState.value.copy(verificationCode = value, errorMessage = null)
    }

    fun onPasswordChange(value: String) {
        _uiState.value = _uiState.value.copy(password = value, errorMessage = null)
    }

    fun onConfirmPasswordChange(value: String) {
        _uiState.value = _uiState.value.copy(confirmPassword = value, errorMessage = null)
    }

    fun onAccountIdChange(value: String) {
        _uiState.value = _uiState.value.copy(accountId = value, errorMessage = null)
    }

    fun onAgreeToTermsChange(value: Boolean) {
        _uiState.value = _uiState.value.copy(agreeToTerms = value)
    }

    fun onReferralCodeChange(value: String) {
        _uiState.value = _uiState.value.copy(referralCode = value, errorMessage = null)
    }

    fun register(onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)

            if (!validateInputs()) {
                return@launch
            }

            val state = _uiState.value
            val request = RegisterRequest(
                phoneNumber = state.phoneNumber,
                accountId = state.accountId,
                password = state.password,
                agreeToTerms = state.agreeToTerms,
                referralCode = state.referralCode.takeIf { it.isNotEmpty() }
            )

            val result = authRepository.register(request)

            if (result.isSuccess) {
                val loginReq = LoginRequest.withPassword(state.phoneNumber, state.password)
                val loginResult = authRepository.login(loginReq)
                if (loginResult.isSuccess) {
                    onSuccess()
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "注册成功，但自动登录失败"
                    )
                }
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.exceptionOrNull()?.message ?: "注册失败"
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

    private fun validateInputs(): Boolean {
        val state = _uiState.value
        when {
            state.phoneNumber.length != 11 -> {
                _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = "请输入正确的手机号")
                return false
            }
            state.password.length < 8 -> {
                _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = "密码至少8位")
                return false
            }
            state.password != state.confirmPassword -> {
                _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = "两次密码不一致")
                return false
            }
            state.accountId.isEmpty() -> {
                _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = "请输入账户ID")
                return false
            }
            !state.agreeToTerms -> {
                _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = "请同意服务条款")
                return false
            }
            else -> return true
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
