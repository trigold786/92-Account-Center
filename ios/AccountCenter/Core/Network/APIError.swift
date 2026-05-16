import Foundation

enum APIError: LocalizedError {
    case networkError(Error)
    case httpError(statusCode: Int, message: String)
    case unauthorized
    case decodingError(Error)
    case unknown(Error)
    
    var errorDescription: String? {
        switch self {
        case .networkError(let error):
            return "网络连接失败: \(error.localizedDescription)"
        case .httpError(_, let message):
            return message
        case .unauthorized:
            return "登录已过期，请重新登录"
        case .decodingError(let error):
            return "数据解析错误: \(error.localizedDescription)"
        case .unknown(let error):
            return "未知错误: \(error.localizedDescription)"
        }
    }
}
