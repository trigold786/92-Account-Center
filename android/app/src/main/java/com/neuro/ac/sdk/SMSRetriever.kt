package com.neuro.ac.sdk

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.Bundle
import com.google.android.gms.auth.api.phone.SmsRetriever
import com.google.android.gms.common.api.CommonStatusCodes
import com.google.android.gms.common.api.Status
import java.security.MessageDigest

class SMSRetriever(private val context: Context) {

    interface SMSCodeListener {
        fun onCodeReceived(code: String)
        fun onTimeout()
    }

    private var listener: SMSCodeListener? = null

    fun start(listener: SMSCodeListener) {
        this.listener = listener
        val client = SmsRetriever.getClient(context)
        val task = client.startSmsRetriever()
        task.addOnSuccessListener {
            val filter = IntentFilter(SmsRetriever.SMS_RETRIEVED_ACTION)
            context.registerReceiver(smsReceiver, filter)
        }
        task.addOnFailureListener {
            listener.onTimeout()
        }
    }

    fun stop() {
        try {
            context.unregisterReceiver(smsReceiver)
        } catch (_: Exception) {
        }
        listener = null
    }

    fun getAppHash(): String {
        try {
            val packageName = context.packageName
            val signature = getAppSignature(packageName)
            val hash = generateHash(packageName, signature)
            return hash.substring(0, 11)
        } catch (_: Exception) {
            return ""
        }
    }

    private val smsReceiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context, intent: Intent) {
            if (SmsRetriever.SMS_RETRIEVED_ACTION == intent.action) {
                val extras = intent.extras
                val status = extras?.get(SmsRetriever.EXTRA_STATUS) as? Status
                when (status?.statusCode) {
                    CommonStatusCodes.SUCCESS -> {
                        val message = extras?.getString(SmsRetriever.EXTRA_SMS_MESSAGE) ?: ""
                        val code = extractCode(message)
                        if (code != null) {
                            listener?.onCodeReceived(code)
                        }
                    }
                    CommonStatusCodes.TIMEOUT -> {
                        listener?.onTimeout()
                    }
                }
            }
        }
    }

    private fun extractCode(message: String): String? {
        val regex = Regex("\\b\\d{4,6}\\b")
        return regex.find(message)?.value
    }

    private fun getAppSignature(packageName: String): String {
        val pm = context.packageManager
        val packageInfo = pm.getPackageInfo(packageName, android.content.pm.PackageManager.GET_SIGNING_CERTIFICATES)
        val signatures = packageInfo.signingInfo?.apkContentsSigners ?: return ""
        val signature = signatures.firstOrNull() ?: return ""
        val md = MessageDigest.getInstance("SHA-256")
        return md.digest(signature.toByteArray()).joinToString("") { "%02x".format(it) }
    }

    private fun generateHash(packageName: String, signature: String): String {
        val input = "$packageName $signature"
        val bytes = input.toByteArray(Charsets.UTF_8)
        val md = MessageDigest.getInstance("SHA-256")
        val hash = md.digest(bytes)
        return android.util.Base64.encodeToString(hash, android.util.Base64.NO_PADDING or android.util.Base64.NO_WRAP)
            .substring(0, 11)
    }
}
