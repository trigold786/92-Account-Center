package com.accountcenter.storage

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import com.accountcenter.model.Token
import com.google.gson.Gson
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "auth")

@Singleton
class TokenManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val gson = Gson()
    private val tokenKey = stringPreferencesKey("token")

    suspend fun saveToken(token: Token) {
        context.dataStore.edit { prefs ->
            prefs[tokenKey] = gson.toJson(token)
        }
    }

    fun getTokenFlow(): Flow<Token?> {
        return context.dataStore.data.map { prefs ->
            prefs[tokenKey]?.let { gson.fromJson(it, Token::class.java) }
        }
    }

    suspend fun getAccessToken(): String? = getTokenFlow().firstOrNull()?.accessToken
    suspend fun getRefreshToken(): String? = getTokenFlow().firstOrNull()?.refreshToken
    suspend fun getToken(): Token? = getTokenFlow().firstOrNull()

    suspend fun clearTokens() {
        context.dataStore.edit { prefs ->
            prefs.remove(tokenKey)
        }
    }
}
