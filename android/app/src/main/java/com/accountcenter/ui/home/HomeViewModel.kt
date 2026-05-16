package com.accountcenter.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.accountcenter.repository.AuthRepository
import com.accountcenter.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    userRepository: UserRepository
) : ViewModel() {
    val userDisplay: StateFlow<UserDisplay> = userRepository.currentUser.map { user ->
        UserDisplay(
            accountId = user?.accountId ?: "",
            phoneNumber = user?.phoneNumber ?: ""
        )
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), UserDisplay())

    fun logout(onLogout: () -> Unit) {
        viewModelScope.launch {
            authRepository.logout()
            onLogout()
        }
    }
}
