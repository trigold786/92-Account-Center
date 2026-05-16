package com.accountcenter.di

import com.accountcenter.network.ApiClient
import com.accountcenter.repository.AuthRepository
import com.accountcenter.repository.UserRepository
import com.accountcenter.storage.TokenManager
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object RepositoryModule {
    @Provides
    @Singleton
    fun provideAuthRepository(apiClient: ApiClient, tokenManager: TokenManager) =
        AuthRepository(apiClient, tokenManager)

    @Provides
    @Singleton
    fun provideUserRepository(tokenManager: TokenManager) =
        UserRepository(tokenManager)
}
