package com.accountcenter.network

sealed class ApiError(open val message: String) {
    data class NetworkError(override val message: String) : ApiError(message)
    data class HttpError(val statusCode: Int, override val message: String) : ApiError(message)
    data object Unauthorized : ApiError("登录已过期，请重新登录")
    data class DecodingError(override val message: String) : ApiError(message)
    data class UnknownError(override val message: String) : ApiError(message)
}
