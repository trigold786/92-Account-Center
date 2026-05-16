import Foundation
import Security

class TokenManager {
    static let shared = TokenManager()
    
    private let service = "com.accountcenter.app"
    private let accessTokenKey = "access_token"
    private let refreshTokenKey = "refresh_token"
    private let tokenDataKey = "token_data"
    
    private init() {}
    
    func save(token: Token) {
        set(token.accessToken, forKey: accessTokenKey)
        set(token.refreshToken, forKey: refreshTokenKey)
        if let jsonData = try? JSONEncoder().encode(token),
           let jsonString = String(data: jsonData, encoding: .utf8) {
            set(jsonString, forKey: tokenDataKey)
        }
    }
    
    func getAccessToken() -> String? {
        get(forKey: accessTokenKey)
    }
    
    func getRefreshToken() -> String? {
        get(forKey: refreshTokenKey)
    }
    
    func getUserId() -> Int64? {
        getTokenData()?.userId
    }
    
    func getAccountId() -> String? {
        getTokenData()?.accountId
    }
    
    private func getTokenData() -> Token? {
        guard let json = get(forKey: tokenDataKey),
              let data = json.data(using: .utf8) else { return nil }
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        return try? decoder.decode(Token.self, from: data)
    }
    
    func clear() {
        delete(forKey: accessTokenKey)
        delete(forKey: refreshTokenKey)
        delete(forKey: tokenDataKey)
    }
    
    private func set(_ value: String, forKey key: String) {
        guard let data = value.data(using: .utf8) else { return }
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
            kSecValueData: data
        ]
        
        SecItemDelete(query as CFDictionary)
        
        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            print("Keychain error: \(status)")
            return
        }
    }
    
    private func get(forKey key: String) -> String? {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
            kSecReturnData: kCFBooleanTrue!,
            kSecMatchLimit: kSecMatchLimitOne
        ]
        
        var dataRef: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &dataRef)
        
        guard status == errSecSuccess,
              let data = dataRef as? Data,
              let value = String(data: data, encoding: .utf8) else {
            return nil
        }
        return value
    }
    
    private func delete(forKey key: String) {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key
        ]
        SecItemDelete(query as CFDictionary)
    }
}
