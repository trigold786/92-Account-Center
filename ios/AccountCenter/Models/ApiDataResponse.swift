import Foundation

struct ApiDataResponse<T: Codable>: Codable {
    let code: Int
    let message: String?
    let data: T?
}
