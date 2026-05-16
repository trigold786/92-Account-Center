package com.accountcenter.repository

import com.accountcenter.network.ApiError
import com.google.gson.Gson

fun Throwable.toApiError(): ApiError = when (this) {
    is retrofit2.HttpException -> {
        val errorBody = response()?.errorBody()?.string()
        val gson = Gson()
        val map = errorBody?.let {
            try {
                gson.fromJson(it, Map::class.java)
            } catch (e: Exception) {
                null
            }
        }
        val message = map?.get("error") as? String ?: map?.get("message") as? String ?: "请求失败"
        ApiError.HttpError(code(), message)
    }
    is java.net.UnknownHostException -> ApiError.NetworkError("网络连接失败")
    is java.net.SocketTimeoutException -> ApiError.NetworkError("请求超时")
    else -> ApiError.UnknownError(message ?: "未知错误")
}
