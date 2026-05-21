package com.neuro.ac.sdk

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import org.json.JSONObject

enum class CachedDataType(val key: String) {
    CREDITS("cached_credits"),
    TIER("cached_tier"),
    SUBSCRIPTION_STATUS("cached_subscription_status"),
    LAST_SYNC("cached_last_sync")
}

class OfflineCacheManager(context: Context) {
    private val appContext = context.applicationContext

    private val masterKey = MasterKey.Builder(appContext)
        .setKeyGenParameterSpec(
            javax.crypto.spec.SecretKeySpec(
                ByteArray(32),
                "AES"
            ).let {
                android.security.keystore.KeyGenParameterSpec.Builder(
                    MasterKey.DEFAULT_MASTER_KEY_ALIAS,
                    android.security.keystore.KeyProperties.PURPOSE_ENCRYPT or
                            android.security.keystore.KeyProperties.PURPOSE_DECRYPT
                )
                    .setBlockModes(android.security.keystore.KeyProperties.BLOCK_MODE_GCM)
                    .setEncryptionPaddings(android.security.keystore.KeyProperties.ENCRYPTION_PADDING_NONE)
                    .setKeySize(256)
                    .build()
            }
        )
        .build()

    private val securePrefs = EncryptedSharedPreferences.create(
        appContext,
        "offline_cache_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    private val _isNetworkAvailable = MutableStateFlow(true)
    val isNetworkAvailable: StateFlow<Boolean> = _isNetworkAvailable

    private var networkCallback: ConnectivityManager.NetworkCallback? = null

    init {
        startNetworkMonitoring()
    }

    fun <T> cache(value: T, type: CachedDataType) {
        val json = when (value) {
            is String -> value
            is Number -> value.toString()
            is Boolean -> value.toString()
            else -> JSONObject().apply { put("value", value.toString()) }.toString()
        }
        securePrefs.edit()
            .putString(type.key, json)
            .putLong(CachedDataType.LAST_SYNC.key, System.currentTimeMillis() / 1000)
            .apply()
    }

    fun getString(type: CachedDataType): String? {
        return securePrefs.getString(type.key, null)
    }

    fun getLong(type: CachedDataType): Long {
        return securePrefs.getLong(type.key, 0L)
    }

    fun getInt(type: CachedDataType): Int {
        return securePrefs.getInt(type.key, 0)
    }

    fun clearCache() {
        CachedDataType.entries.forEach { type ->
            securePrefs.edit().remove(type.key).apply()
        }
    }

    fun isCacheValid(maxAgeSeconds: Int = 3600): Boolean {
        val lastSync = securePrefs.getLong(CachedDataType.LAST_SYNC.key, 0L)
        val now = System.currentTimeMillis() / 1000
        return (now - lastSync) < maxAgeSeconds
    }

    private fun startNetworkMonitoring() {
        val cm = appContext.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()

        val callback = object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) {
                val wasOffline = !_isNetworkAvailable.value
                _isNetworkAvailable.value = true
                if (wasOffline) {
                    onNetworkRestored()
                }
            }

            override fun onLost(network: Network) {
                _isNetworkAvailable.value = false
            }
        }
        cm.registerNetworkCallback(request, callback)
        networkCallback = callback
    }

    private fun onNetworkRestored() {
        syncFromServer()
    }

    private fun syncFromServer() {
        val serverData = fetchServerData()
        serverData?.forEach { (type, value) ->
            cache(value, type)
        }
    }

    private fun fetchServerData(): Map<CachedDataType, Any>? {
        return null
    }

    fun resolveConflict(localValue: Any?, serverValue: Any?): Any? {
        return serverValue
    }

    fun destroy() {
        networkCallback?.let {
            val cm = appContext.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
            cm.unregisterNetworkCallback(it)
        }
    }
}
