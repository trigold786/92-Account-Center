import Foundation

class APIClient {
    static let shared = APIClient()
    
    private let session: URLSession
    
    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: config)
    }
    
    private func request<T: Decodable>(
        method: String,
        url: URL,
        body: Encodable? = nil,
        token: String? = nil
    ) async throws -> T {
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let token = token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }
        
        do {
            let (data, response) = try await session.data(for: request)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                throw APIError.unknown(NSError(domain: "APIClient", code: -1))
            }
            
            switch httpResponse.statusCode {
            case 200...299:
                return try decodeResponse(data: data)
            case 401:
                // Attempt token refresh and retry once
                if await AuthManager.shared.refreshIfNeeded() {
                    var retryRequest = URLRequest(url: url)
                    retryRequest.httpMethod = method
                    retryRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
                    if let newToken = AuthManager.shared.accessToken {
                        retryRequest.setValue("Bearer \(newToken)", forHTTPHeaderField: "Authorization")
                    }
                    if let body = body {
                        retryRequest.httpBody = try JSONEncoder().encode(body)
                    }
                    let (retryData, retryResponse) = try await session.data(for: retryRequest)
                    guard let retryHttpResponse = retryResponse as? HTTPURLResponse else {
                        throw APIError.unknown(NSError(domain: "APIClient", code: -1))
                    }
                    if retryHttpResponse.statusCode == 401 {
                        throw APIError.unauthorized
                    }
                    return try decodeResponse(data: retryData)
                }
                throw APIError.unauthorized
            default:
                let decoder = JSONDecoder()
                let errorResponse = try? decoder.decode([String: String].self, from: data)
                let message = errorResponse?["error"] ?? errorResponse?["message"] ?? "请求失败"
                throw APIError.httpError(statusCode: httpResponse.statusCode, message: message)
            }
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.networkError(error)
        }
    }
    
    private func decodeResponse<T: Decodable>(data: Data) throws -> T {
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ss'Z'"
        dateFormatter.timeZone = TimeZone(secondsFromGMT: 0)
        decoder.dateDecodingStrategy = .formatted(dateFormatter)
        return try decoder.decode(T.self, from: data)
    }
    
    func login(request: LoginRequest) async throws -> Token {
        return try await request(method: "POST", url: Endpoints.login, body: request)
    }
    
    func refresh(refreshToken: String) async throws -> Token {
        let request = RefreshTokenRequest(refreshToken: refreshToken)
        return try await request(method: "POST", url: Endpoints.refresh, body: request)
    }
    
    func logout(accessToken: String) async throws {
        let _: [String: String] = try await request(method: "POST", url: Endpoints.logout, token: accessToken)
    }
    
    func sendSMS(request: SMSSendRequest) async throws {
        let _: [String: String] = try await request(method: "POST", url: Endpoints.sendSMS, body: request)
    }
    
    func register(request: RegisterRequest) async throws -> RegisterResponse {
        return try await request(method: "POST", url: Endpoints.register, body: request)
    }
}
