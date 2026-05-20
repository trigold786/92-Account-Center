import Foundation

enum AppStrings {
    enum Common {
        static let login = String(localized: "common.login", defaultValue: "Login")
        static let register = String(localized: "common.register", defaultValue: "Register")
        static let settings = String(localized: "common.settings", defaultValue: "Settings")
        static let logout = String(localized: "common.logout", defaultValue: "Logout")
        static let save = String(localized: "common.save", defaultValue: "Save")
        static let cancel = String(localized: "common.cancel", defaultValue: "Cancel")
    }

    enum Credits {
        static let balance = String(localized: "credits.balance", defaultValue: "Credits Balance")
        static let earn = String(localized: "credits.earn", defaultValue: "Earn Credits")
        static let history = String(localized: "credits.history", defaultValue: "Credit History")
        static let referral = String(localized: "credits.referral", defaultValue: "Referral Rewards")
    }

    enum Subscription {
        static let plan = String(localized: "subscription.plan", defaultValue: "Subscription Plan")
        static let upgrade = String(localized: "subscription.upgrade", defaultValue: "Upgrade")
        static let downgrade = String(localized: "subscription.downgrade", defaultValue: "Downgrade")
        static let renew = String(localized: "subscription.renew", defaultValue: "Renew")
        static let expire = String(localized: "subscription.expire", defaultValue: "Expire")
    }

    enum Dashboard {
        static let welcome = String(localized: "dashboard.welcome", defaultValue: "Welcome Back")
        static let overview = String(localized: "dashboard.overview", defaultValue: "Overview")
        static let quickActions = String(localized: "dashboard.quick_actions", defaultValue: "Quick Actions")
    }

    enum Errors {
        static let network = String(localized: "errors.network", defaultValue: "Network connection failed.")
        static let server = String(localized: "errors.server", defaultValue: "Server error. Please try again later.")
        static let unauthorized = String(localized: "errors.unauthorized", defaultValue: "Session expired. Please log in again.")
        static let notFound = String(localized: "errors.not_found", defaultValue: "The requested resource was not found.")
    }

    static var currentLanguage: String {
        Locale.current.language.languageCode?.identifier ?? "en"
    }

    static var isChinese: Bool {
        currentLanguage.hasPrefix("zh")
    }
}
