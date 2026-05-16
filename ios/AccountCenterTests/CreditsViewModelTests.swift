import XCTest
@testable import AccountCenter

final class CreditsViewModelTests: XCTestCase {

    func testDecodeCreditAccount() {
        let json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "user_id": 42,
                "balance": 1500.50,
                "status": "active"
            }
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let response = try? decoder.decode(ApiDataResponse<CreditAccount>.self, from: json)
        XCTAssertNotNil(response)
        XCTAssertEqual(response?.code, 200)
        XCTAssertEqual(response?.data?.balance, 1500.50)
        XCTAssertEqual(response?.data?.status, "active")
    }

    func testDecodeTransactionList() {
        let json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "transactions": [
                    {"id":1,"credit_account_id":10,"type":"earn","amount":100.0,"reference_id":"ref_1","details":"邀请奖励","status":"completed","created_at":"2026-05-01T10:00:00Z"},
                    {"id":2,"credit_account_id":10,"type":"consume","amount":50.0,"reference_id":"ref_2","details":"积分消费","status":"completed","created_at":"2026-05-02T10:00:00Z"}
                ],
                "total": 2,
                "page": 1,
                "page_size": 20
            }
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let response = try? decoder.decode(ApiDataResponse<TransactionList>.self, from: json)
        XCTAssertNotNil(response)
        XCTAssertEqual(response?.data?.transactions.count, 2)
        XCTAssertEqual(response?.data?.total, 2)
        XCTAssertEqual(response?.data?.transactions.first?.type, "earn")
        XCTAssertEqual(response?.data?.transactions.last?.amount, 50.0)
    }

    func testDecodeReferralSummary() {
        let json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "total_referees": 10,
                "total_earned": 500.0,
                "active_referees": 5
            }
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let response = try? decoder.decode(ApiDataResponse<ReferralSummary>.self, from: json)
        XCTAssertNotNil(response)
        XCTAssertEqual(response?.data?.totalReferees, 10)
        XCTAssertEqual(response?.data?.totalEarned, 500.0)
        XCTAssertEqual(response?.data?.activeReferees, 5)
    }

    func testTransactionAmountDisplay() {
        let json = """
        [
            {"id":1,"credit_account_id":10,"type":"earn","amount":100.0,"details":"邀请奖励","status":"completed","created_at":"2026-05-01T10:00:00Z"},
            {"id":2,"credit_account_id":10,"type":"consume","amount":50.0,"details":"积分消费","status":"completed","created_at":"2026-05-02T10:00:00Z"}
        ]
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let txs = try! decoder.decode([Transaction].self, from: json)
        XCTAssertEqual(txs[0].type, "earn")
        XCTAssertEqual(txs[1].type, "consume")
    }
}
