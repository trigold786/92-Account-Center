import Foundation
import Security

class TokenManager {
    static let shared = TokenManager()
    
    private let service = "com.accountcenter.app"
    private let accessTokenKey = "access_token"
    private let refreshTokenKey = "refresh_token"
    
    private init() {}
    
    func save(token: Token) {
        set(token.accessToken, forKey: accessTokenKey)
        set(token.refreshToken, forKey: refreshTokenKey)
    }
    
    func getAccessToken() -> String? {
        get(forKey: accessTokenKey)
    }
    
    func getRefreshToken() -> String? {
        get(forKey: refreshTokenKey)
    }
    
    func clear() {
        delete(forKey: accessTokenKey)
        delete(forKey: refreshTokenKey)
    }
    
    private func set(_ value: String, forKey key: String) {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
            kSecValueData: value.data(using: .utf8)!
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
