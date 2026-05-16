import XCTest
@testable import AccountCenter

final class SubscriptionViewModelTests: XCTestCase {

    func testDecodeSubscription() {
        let json = """
        {
            "id": 1,
            "user_id": 42,
            "tier_level": 2,
            "start_time": "2026-01-01T00:00:00Z",
            "end_time": "2026-12-31T23:59:59Z",
            "status": "active",
            "price": 99.99,
            "payment_method": "alipay",
            "order_id": "ORD_123"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let sub = try? decoder.decode(Subscription.self, from: json)
        XCTAssertNotNil(sub)
        XCTAssertEqual(sub?.id, 1)
        XCTAssertEqual(sub?.userId, 42)
        XCTAssertEqual(sub?.tierLevel, 2)
        XCTAssertEqual(sub?.status, "active")
        XCTAssertEqual(sub?.price, 99.99)
    }

    func testDecodeSubscriptionList() {
        let json = """
        [
            {"id":1,"user_id":42,"tier_level":1,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-06-30T23:59:59Z","status":"expired","price":29.99},
            {"id":2,"user_id":42,"tier_level":2,"start_time":"2026-07-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z","status":"active","price":99.99}
        ]
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let subs = try? decoder.decode([Subscription].self, from: json)
        XCTAssertNotNil(subs)
        XCTAssertEqual(subs?.count, 2)
        XCTAssertEqual(subs?.first?.status, "expired")
        XCTAssertEqual(subs?.last?.status, "active")
    }

    func testDecodeTierInfo() {
        let json = """
        {
            "user_id": 42,
            "identity_tier": 2
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let tier = try? decoder.decode(TierInfo.self, from: json)
        XCTAssertNotNil(tier)
        XCTAssertEqual(tier?.userId, 42)
        XCTAssertEqual(tier?.identityTier, 2)
    }

    func testActiveSubscriptionsFilter() {
        let json = """
        [
            {"id":1,"user_id":42,"tier_level":1,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-06-30T23:59:59Z","status":"expired","price":29.99},
            {"id":2,"user_id":42,"tier_level":2,"start_time":"2026-07-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z","status":"active","price":99.99}
        ]
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let subs = try! decoder.decode([Subscription].self, from: json)
        let active = subs.filter { $0.status == "active" }
        XCTAssertEqual(active.count, 1)
        XCTAssertEqual(active.first?.tierLevel, 2)
    }
}
