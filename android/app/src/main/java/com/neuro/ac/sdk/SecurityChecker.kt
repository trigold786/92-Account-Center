package com.neuro.ac.sdk

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.util.Log
import java.io.File

enum class SecurityLevel {
    SECURE,
    COMPROMISED,
    UNKNOWN
}

class SecurityChecker(private val context: Context) {
    companion object {
        private const val TAG = "SecurityChecker"
    }

    private val suPaths = listOf(
        "/system/bin/su",
        "/system/xbin/su",
        "/sbin/su",
        "/data/local/xbin/su",
        "/data/local/bin/su",
        "/system/sd/xbin/su",
        "/system/bin/failsafe/su",
        "/su/bin/su",
        "/data/local/su",
        "/magisk/.core/bin/su"
    )

    private val suspiciousFiles = listOf(
        "/system/app/Superuser.apk",
        "/system/app/SuperSU.apk",
        "/system/app/Magisk.apk",
        "/sbin/.magisk",
        "/cache/.disable_magisk",
        "/dev/.magisk.unblock",
        "/data/adb/magisk"
    )

    private val suspiciousPackages = listOf(
        "com.topjohnwu.magisk",
        "eu.chainfire.supersu",
        "com.noshufou.android.su",
        "com.koushikdutta.superuser",
        "me.phh.superuser",
        "com.thirdparty.superuser",
        "com.yellowes.su",
        "com.koushikdutta.rommanager",
        "com.dimonvideo.luckypatcher",
        "com.chelpus.lackypatch"
    )

    fun checkSecurity(): SecurityLevel {
        var isCompromised = false

        if (checkSuBinary()) isCompromised = true
        if (!isCompromised && checkSuspiciousFiles()) isCompromised = true
        if (!isCompromised && checkSuspiciousPackages()) isCompromised = true
        if (!isCompromised && checkBuildTags()) isCompromised = true

        val level = if (isCompromised) SecurityLevel.COMPROMISED else SecurityLevel.SECURE
        reportToBackend(level)
        return level
    }

    private fun checkSuBinary(): Boolean {
        return suPaths.any { path -> File(path).exists() }
    }

    private fun checkSuspiciousFiles(): Boolean {
        return suspiciousFiles.any { path -> File(path).exists() }
    }

    private fun checkSuspiciousPackages(): Boolean {
        val pm = context.packageManager
        return suspiciousPackages.any { pkg ->
            try {
                pm.getPackageInfo(pkg, 0)
                true
            } catch (_: PackageManager.NameNotFoundException) {
                false
            }
        }
    }

    private fun checkBuildTags(): Boolean {
        return Build.TAGS?.contains("test-keys") == true
    }

    private fun reportToBackend(level: SecurityLevel) {
        Thread {
            try {
                val payload = org.json.JSONObject().apply {
                    put("security_level", level.name.lowercase())
                    put("timestamp", System.currentTimeMillis() / 1000)
                    put("platform", "android")
                    put("device_model", Build.MODEL)
                    put("android_version", Build.VERSION.SDK_INT)
                }
                val url = java.net.URL("https://ac.neuro.ai/api/v1/security/report")
                val connection = url.openConnection() as java.net.HttpURLConnection
                connection.requestMethod = "POST"
                connection.setRequestProperty("Content-Type", "application/json")
                connection.doOutput = true
                connection.outputStream.use { os ->
                    os.write(payload.toString().toByteArray(Charsets.UTF_8))
                }
                connection.responseCode
                connection.disconnect()
            } catch (e: Exception) {
                Log.e(TAG, "Failed to report security status", e)
            }
        }.start()
    }
}
