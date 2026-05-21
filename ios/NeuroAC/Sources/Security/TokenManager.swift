import Foundation
import Security
import CryptoKit

class NeuroTokenManager {
    static let shared = NeuroTokenManager()
    private let accessTokenKey = "com.neuro.ac.accessToken"
    private let refreshTokenKey = "com.neuro.ac.refreshToken"
    private let fingerprintKey = "com.neuro.ac.deviceFingerprint"
    private let keychainService = "com.neuro.ac.keychain"

    private func saveToKeychain(key: String, value: String) {
        guard let data = value.data(using: .utf8) else { return }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    private func loadFromKeychain(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    private func deleteFromKeychain(key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(query as CFDictionary)
    }

    var accessToken: String? {
        get { loadFromKeychain(key: accessTokenKey) }
        set {
            if let value = newValue {
                saveToKeychain(key: accessTokenKey, value: value)
            } else {
                deleteFromKeychain(key: accessTokenKey)
            }
        }
    }

    var refreshToken: String? {
        get { loadFromKeychain(key: refreshTokenKey) }
        set {
            if let value = newValue {
                saveToKeychain(key: refreshTokenKey, value: value)
            } else {
                deleteFromKeychain(key: refreshTokenKey)
            }
        }
    }

    func generateDeviceFingerprint() -> String {
        var bytes = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        let fingerprint = bytes.map { String(format: "%02x", $0) }.joined()
        saveToKeychain(key: fingerprintKey, value: fingerprint)
        return fingerprint
    }

    var deviceFingerprint: String? {
        loadFromKeychain(key: fingerprintKey)
    }

    func clearAll() {
        deleteFromKeychain(key: accessTokenKey)
        deleteFromKeychain(key: refreshTokenKey)
        deleteFromKeychain(key: fingerprintKey)
    }
}
