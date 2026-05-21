import UIKit

class ScreenshotProtection {
    static let shared = ScreenshotProtection()

    private var isMonitoring = false
    private var onScreenshotDetected: (() -> Void)?
    private var protectedWindows: [UIWindow] = []

    func startMonitoring(onDetected: @escaping () -> Void) {
        guard !isMonitoring else { return }
        isMonitoring = true
        onScreenshotDetected = onDetected

        NotificationCenter.default.addObserver(
            self,
            selector: #selector(screenCapturedChanged),
            name: UIScreen.capturedDidChangeNotification,
            object: nil
        )

        NotificationCenter.default.addObserver(
            self,
            selector: #selector(userDidTakeScreenshot),
            name: UIApplication.userDidTakeScreenshotNotification,
            object: nil
        )
    }

    func stopMonitoring() {
        guard isMonitoring else { return }
        isMonitoring = false
        NotificationCenter.default.removeObserver(self)
        onScreenshotDetected = nil
    }

    @objc private func screenCapturedChanged() {
        if UIScreen.main.isCaptured {
            onScreenshotDetected?()
            showPrivacyOverlay()
        } else {
            hidePrivacyOverlay()
        }
    }

    @objc private func userDidTakeScreenshot() {
        onScreenshotDetected?()
        showPrivacyAlert()
    }

    func setSecureFlag(for window: UIWindow) {
        protectedWindows.append(window)
        if #available(iOS 17.0, *) {
            window.screen.isCaptured
        }
        observeScreenCapture(for: window)
    }

    private func observeScreenCapture(for window: UIWindow) {
        if UIScreen.main.isCaptured {
            window.layer.opacity = 0
        }
    }

    private var overlayWindow: UIWindow?

    private func showPrivacyOverlay() {
        guard overlayWindow == nil else { return }
        let scenes = UIApplication.shared.connectedScenes
        let windowScene = scenes.first as? UIWindowScene

        let overlay = UIWindow(windowScene: windowScene!)
        overlay.frame = UIScreen.main.bounds
        overlay.backgroundColor = .black
        overlay.windowLevel = .statusBar + 1

        let label = UILabel()
        label.text = "Screen recording detected"
        label.textColor = .white
        label.textAlignment = .center
        label.font = .systemFont(ofSize: 20, weight: .semibold)
        label.translatesAutoresizingMaskIntoConstraints = false
        overlay.addSubview(label)
        NSLayoutConstraint.activate([
            label.centerXAnchor.constraint(equalTo: overlay.centerXAnchor),
            label.centerYAnchor.constraint(equalTo: overlay.centerYAnchor),
        ])

        overlay.isHidden = false
        overlayWindow = overlay
    }

    private func hidePrivacyOverlay() {
        overlayWindow?.isHidden = true
        overlayWindow = nil
    }

    private func showPrivacyAlert() {
        let scenes = UIApplication.shared.connectedScenes
        let windowScene = scenes.first as? UIWindowScene
        guard let rootVC = windowScene?.windows.first?.rootViewController else { return }

        var topVC = rootVC
        while let presented = topVC.presentedViewController {
            topVC = presented
        }

        let alert = UIAlertController(
            title: "Screenshot Detected",
            message: "This screen contains sensitive information. Please be careful when sharing screenshots.",
            preferredStyle: .alert
        )
        alert.addAction(UIAlertAction(title: "OK", style: .default))
        topVC.present(alert, animated: true)
    }

    deinit {
        NotificationCenter.default.removeObserver(self)
    }
}
