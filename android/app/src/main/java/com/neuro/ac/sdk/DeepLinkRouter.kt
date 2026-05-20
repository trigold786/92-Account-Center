package com.neuro.ac.sdk

import android.content.Intent
import android.net.Uri

sealed class DeepLinkDestination {
    data class Subscription(val tier: String?) : DeepLinkDestination()
    data class ReferralRegister(val inviterId: String) : DeepLinkDestination()
    data object Credits : DeepLinkDestination()
    data object Home : DeepLinkDestination()
}

class DeepLinkRouter {
    companion object {
        private const val CUSTOM_SCHEME = "neuro"
        private const val UNIVERSAL_HOST = "ac.neuro.ai"
    }

    fun route(intent: Intent): DeepLinkDestination? {
        val data = intent.data ?: return null
        return route(data)
    }

    fun route(uri: Uri): DeepLinkDestination? {
        return when (uri.scheme) {
            CUSTOM_SCHEME -> parseCustomScheme(uri)
            "https" -> if (uri.host == UNIVERSAL_HOST) parseUniversalLink(uri) else null
            "http" -> if (uri.host == UNIVERSAL_HOST) parseUniversalLink(uri) else null
            else -> null
        }
    }

    private fun parseCustomScheme(uri: Uri): DeepLinkDestination {
        val host = uri.host ?: return DeepLinkDestination.Home
        val params = extractQueryParams(uri)

        return when (host) {
            "subscribe" -> DeepLinkDestination.Subscription(tier = params["tier"])
            "register" -> {
                val inviterId = params["inviter_id"]
                if (inviterId != null) DeepLinkDestination.ReferralRegister(inviterId)
                else DeepLinkDestination.Home
            }
            "referral" -> {
                val inviterId = params["inviter_id"]
                if (inviterId != null) DeepLinkDestination.ReferralRegister(inviterId)
                else DeepLinkDestination.Home
            }
            "credits" -> DeepLinkDestination.Credits
            else -> DeepLinkDestination.Home
        }
    }

    private fun parseUniversalLink(uri: Uri): DeepLinkDestination {
        val segments = uri.pathSegments
        if (segments.isEmpty()) return DeepLinkDestination.Home
        val params = extractQueryParams(uri)

        return when (segments[0]) {
            "subscribe" -> {
                val tier = params["tier"] ?: segments.getOrNull(1)
                DeepLinkDestination.Subscription(tier = tier)
            }
            "register" -> {
                val inviterId = params["inviter_id"]
                if (inviterId != null) DeepLinkDestination.ReferralRegister(inviterId)
                else DeepLinkDestination.Home
            }
            "referral" -> {
                val inviterId = params["inviter_id"]
                if (inviterId != null) DeepLinkDestination.ReferralRegister(inviterId)
                else DeepLinkDestination.Home
            }
            "credits" -> DeepLinkDestination.Credits
            else -> DeepLinkDestination.Home
        }
    }

    private fun extractQueryParams(uri: Uri): Map<String, String> {
        val params = mutableMapOf<String, String>()
        for (paramName in uri.queryParameterNames) {
            uri.getQueryParameter(paramName)?.let { value ->
                params[paramName] = value
            }
        }
        return params
    }
}
