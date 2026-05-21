package com.neuro.ac.sdk

import android.app.Activity
import android.app.AlertDialog
import android.content.ContentResolver
import android.database.ContentObserver
import android.net.Uri
import android.os.Handler
import android.os.Looper
import android.provider.MediaStore
import android.view.WindowManager

class ScreenshotProtection {
    companion object {
        private const val TAG = "ScreenshotProtection"
    }

    private var contentObserver: ContentObserver? = null
    private var isMonitoring = false
    private var onScreenshotDetected: (() -> Unit)? = null

    fun applySecureFlag(activity: Activity) {
        activity.window.setFlags(
            WindowManager.LayoutParams.FLAG_SECURE,
            WindowManager.LayoutParams.FLAG_SECURE
        )
    }

    fun removeSecureFlag(activity: Activity) {
        activity.window.clearFlags(WindowManager.LayoutParams.FLAG_SECURE)
    }

    fun startMonitoring(activity: Activity, onDetected: () -> Unit) {
        if (isMonitoring) return
        isMonitoring = true
        onScreenshotDetected = onDetected

        val handler = Handler(Looper.getMainLooper())
        contentObserver = object : ContentObserver(handler) {
            override fun onChange(selfChange: Boolean) {
                super.onChange(selfChange)
                onScreenshotDetected?.invoke()
                showPrivacyAlert(activity)
            }
        }

        activity.contentResolver.registerContentObserver(
            MediaStore.Images.Media.EXTERNAL_CONTENT_URI,
            true,
            contentObserver!!
        )
    }

    fun stopMonitoring(activity: Activity) {
        if (!isMonitoring) return
        isMonitoring = false
        contentObserver?.let {
            activity.contentResolver.unregisterContentObserver(it)
        }
        contentObserver = null
        onScreenshotDetected = null
    }

    fun showPrivacyOverlay(activity: Activity) {
        activity.runOnUiThread {
            val dialog = AlertDialog.Builder(activity)
                .setTitle("Screen Recording Detected")
                .setMessage("This screen contains sensitive information.")
                .setPositiveButton("OK", null)
                .setCancelable(false)
                .create()
            dialog.window?.setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY)
            dialog.show()
        }
    }

    private fun showPrivacyAlert(activity: Activity) {
        activity.runOnUiThread {
            val dialog = AlertDialog.Builder(activity)
                .setTitle("Screenshot Detected")
                .setMessage("This screen contains sensitive information. Please be careful when sharing screenshots.")
                .setPositiveButton("OK", null)
                .create()
            dialog.show()
        }
    }
}
