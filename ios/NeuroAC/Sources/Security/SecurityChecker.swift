import Foundation
import UIKit

enum SecurityLevel: String {
    case secure
    case compromised
    case unknown
}

class SecurityChecker {
    static let shared = SecurityChecker()

    private let suspiciousPaths = [
        "/Applications/Cydia.app",
        "/Applications/SBSettings.app",
        "/Applications/Icy.app",
        "/Applications/IntelliScreen.app",
        "/Applications/Snoop-it-config.app",
        "/Library/MobileSubstrate/MobileSubstrate.dylib",
        "/usr/sbin/sshd",
        "/usr/bin/sshd",
        "/usr/libexec/sftp-server",
        "/private/var/lib/apt",
        "/private/var/lib/cydia",
        "/private/var/tmp/cydia.log",
        "/usr/bin/cycript",
        "/usr/local/bin/cycript",
        "/usr/lib/libcycript.dylib",
        "/bin/bash",
        "/bin/sh",
        "/usr/sbin/frida-server",
        "/usr/lib/libfrida.dylib"
    ]

    private let suspiciousDylibs = [
        "SubstrateLoader.dylib",
        "SSLKillSwitch2.dylib",
        "TrustMe.dylib",
        "cycript"
    ]

    func checkSecurity() -> SecurityLevel {
        var isCompromised = false

        for path in suspiciousPaths {
            if FileManager.default.fileExists(atPath: path) {
                isCompromised = true
                break
            }
        }

        if !isCompromised {
            isCompromised = checkSandboxIntegrity()
        }

        if !isCompromised {
            isCompromised = checkDylibs()
        }

        let level: SecurityLevel = isCompromised ? .compromised : .secure
        reportToBackend(level: level)
        return level
    }

    private func checkSandboxIntegrity() -> Bool {
        let testPath = "/private/jailbreak_test_\(UUID().uuidString)"
        do {
            try "test".write(toFile: testPath, atomically: true, encoding: .utf8)
            try? FileManager.default.removeItem(atPath: testPath)
            return true
        } catch {
            return false
        }
    }

    private func checkDylibs() -> Bool {
        let env = ProcessInfo.processInfo.environment
        for dylib in suspiciousDylibs {
            for (_, value) in env where value.contains(dylib) {
                return true
            }
        }

        let dyldInsertLibraries = env["DYLD_INSERT_LIBRARIES"] ?? ""
        if !dyldInsertLibraries.isEmpty {
            return true
        }

        return false
    }

    func canOpenCydiaURL() -> Bool {
        guard let url = URL(string: "cydia://package/com.example.package") else { return false }
        return UIApplication.shared.canOpenURL(url)
    }

    private func reportToBackend(level: SecurityLevel) {
        let payload: [String: Any] = [
            "security_level": level.rawValue,
            "timestamp": Int(Date().timeIntervalSince1970),
            "platform": "ios",
            "device_model": UIDevice.current.model
        ]
        let url = URL(string: "https://ac.neuro.ai/api/v1/security/report")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONSerialization.data(withJSONObject: payload)
        URLSession.shared.dataTask(with: request).resume()
    }
}
