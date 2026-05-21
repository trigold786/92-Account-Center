import Foundation
import CryptoKit
import Network

enum CachedDataType: String {
    case credits = "cached_credits"
    case tier = "cached_tier"
    case subscriptionStatus = "cached_subscription_status"
    case lastSync = "cached_last_sync"
}

class OfflineCacheManager {
    static let shared = OfflineCacheManager()
    private let defaults = UserDefaults.standard
    private let encryptionKey: SymmetricKey
    private var monitor: NWPathMonitor?
    private var isNetworkAvailable = true

    private init() {
        let keyData = Self.getOrCreateKey()
        encryptionKey = SymmetricKey(data: keyData)
        startNetworkMonitoring()
    }

    private static func getOrCreateKey() -> Data {
        let keyKey = "com.neuro.ac.cache.encryption.key"
        if let existing = KeychainHelper.load(key: keyKey) {
            return existing
        }
        let key = SymmetricKey(size: .bits256)
        let data = key.withUnsafeBytes { Data($0) }
        KeychainHelper.save(key: keyKey, data: data)
        return data
    }

    func cache<T: Codable>(_ value: T, for type: CachedDataType) {
        do {
            let encoded = try JSONEncoder().encode(value)
            let sealed = try AES.GCM.seal(encoded, using: encryptionKey)
            let combined = sealed.combined!
            defaults.set(combined.base64EncodedString(), forKey: type.rawValue)
            let now = Int(Date().timeIntervalSince1970)
            defaults.set(now, forKey: CachedDataType.lastSync.rawValue)
        } catch {
            print("Failed to cache \(type.rawValue): \(error)")
        }
    }

    func load<T: Codable>(_ type: CachedDataType, as: T.Type) -> T? {
        guard let base64 = defaults.string(forKey: type.rawValue),
              let data = Data(base64Encoded: base64) else { return nil }
        do {
            let sealed = try AES.GCM.SealedBox(combined: data)
            let decrypted = try AES.GCM.open(sealed, using: encryptionKey)
            return try JSONDecoder().decode(T.self, from: decrypted)
        } catch {
            return nil
        }
    }

    func clearCache() {
        for type in [CachedDataType.credits, .tier, .subscriptionStatus, .lastSync] {
            defaults.removeObject(forKey: type.rawValue)
        }
    }

    func isCacheValid(maxAgeSeconds: Int = 3600) -> Bool {
        let lastSync = defaults.integer(forKey: CachedDataType.lastSync.rawValue)
        let now = Int(Date().timeIntervalSince1970)
        return (now - lastSync) < maxAgeSeconds
    }

    private func startNetworkMonitoring() {
        monitor = NWPathMonitor()
        monitor?.pathUpdateHandler = { [weak self] path in
            let wasOffline = !(self?.isNetworkAvailable ?? true)
            self?.isNetworkAvailable = path.status == .satisfied
            if wasOffline && (self?.isNetworkAvailable ?? false) {
                self?.onNetworkRestored()
            }
        }
        let queue = DispatchQueue(label: "com.neuro.ac.network.monitor")
        monitor?.start(queue: queue)
    }

    private func onNetworkRestored() {
        NotificationCenter.default.post(name: .networkRestored, forKey: "isOnline", object: nil)
    }

    func getNetworkStatus() -> Bool {
        return isNetworkAvailable
    }

    deinit {
        monitor?.cancel()
    }
}

extension Notification.Name {
    static let networkRestored = Notification.Name("com.neuro.ac.network.restored")
}

private struct KeychainHelper {
    static func save(key: String, data: Data) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: "com.neuro.ac.cache",
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    static func load(key: String) -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: "com.neuro.ac.cache",
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess else { return nil }
        return result as? Data
    }
}
