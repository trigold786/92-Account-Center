package com.accountcenter.repository

import com.accountcenter.model.Token
import com.accountcenter.model.User
import com.accountcenter.storage.TokenManager
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    private val tokenManager: TokenManager
) {
    val currentUser: Flow<User?> = tokenManager.getTokenFlow().map { token ->
        token?.let {
            User(
                id = it.userId,
                phoneNumber = null,
                accountId = it.accountId
            )
        }
    }

    suspend fun getToken(): Token? = tokenManager.getToken()
}
