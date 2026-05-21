package com.neuro.ac.sdk

import okhttp3.CertificatePinner
import okhttp3.OkHttpClient

class CertificatePinnerManager {
    companion object {
        private val pins = setOf(
            "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
            "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="
        )

        fun createPinnedClient(): OkHttpClient {
            val certificatePinner = CertificatePinner.Builder()
            pins.forEach { pin ->
                certificatePinner.add("api.neuro.com", pin)
            }
            return OkHttpClient.Builder()
                .certificatePinner(certificatePinner.build())
                .build()
        }
    }
}
