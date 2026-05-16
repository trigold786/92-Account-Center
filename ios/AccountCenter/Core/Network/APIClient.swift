import Foundation

class APIClient {
    static let shared = APIClient()

    private let session: URLSession

    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: config)
    }

    // MARK: - Core request helpers

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

    private func requestWrapped<T: Decodable>(
        method: String,
        url: URL,
        body: Encodable? = nil,
        token: String? = nil
    ) async throws -> T {
        let response: ApiDataResponse<T> = try await request(method: method, url: url, body: body, token: token)
        guard response.code == 200, let data = response.data else {
            throw APIError.httpError(statusCode: response.code, message: response.message ?? "请求失败")
        }
        return data
    }

    private func authRequest<T: Decodable>(method: String, url: URL, body: Encodable? = nil) async throws -> T {
        guard let token = AuthManager.shared.accessToken else { throw APIError.unauthorized }
        return try await request(method: method, url: url, body: body, token: token)
    }

    private func authRequestWrapped<T: Decodable>(method: String, url: URL, body: Encodable? = nil) async throws -> T {
        guard let token = AuthManager.shared.accessToken else { throw APIError.unauthorized }
        return try await requestWrapped(method: method, url: url, body: body, token: token)
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

    // MARK: - Auth

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

    // MARK: - Subscriptions

    func getUserSubscriptions(userId: Int64) async throws -> [Subscription] {
        return try await authRequest(method: "GET", url: Endpoints.userSubscriptions(userId: userId))
    }

    func getUserTier(userId: Int64) async throws -> TierInfo {
        return try await authRequest(method: "GET", url: Endpoints.userTier(userId: userId))
    }

    // MARK: - Credits

    func getCreditAccount(userId: Int64) async throws -> CreditAccount {
        return try await authRequestWrapped(method: "GET", url: Endpoints.creditAccount(userId: userId))
    }

    func getTransactions(userId: Int64, page: Int = 1, pageSize: Int = 20) async throws -> TransactionList {
        return try await authRequestWrapped(method: "GET", url: Endpoints.transactions(userId: userId, page: page, pageSize: pageSize))
    }

    func getReferralSummary(userId: Int64) async throws -> ReferralSummary {
        return try await authRequestWrapped(method: "GET", url: Endpoints.referralSummary(userId: userId))
    }

    func generateReferralLink(userId: Int64) async throws -> ReferralLinkData {
        let body = GenerateLinkRequest(userId: "\(userId)")
        return try await authRequestWrapped(method: "POST", url: Endpoints.generateReferralLink, body: body)
    }

    // MARK: - Risk & Security

    func getRiskHistory(userId: Int64) async throws -> RiskHistoryData {
        return try await authRequestWrapped(method: "GET", url: Endpoints.riskHistory(userId: userId))
    }

    func getUserDevices(userId: Int64) async throws -> DeviceList {
        return try await authRequestWrapped(method: "GET", url: Endpoints.userDevices(userId: userId))
    }

    // MARK: - RFM

    func getRFMScore(userId: Int64) async throws -> RFMScore {
        return try await authRequest(method: "GET", url: Endpoints.rfmScore(userId: userId))
    }
}
