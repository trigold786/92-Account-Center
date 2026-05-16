import XCTest
@testable import AccountCenter

final class APIClientTests: XCTestCase {
    var sut: APIClient!

    override func setUp() {
        super.setUp()
        sut = APIClient.shared
    }

    override func tearDown() {
        sut = nil
        super.tearDown()
    }

    func testDecodeResponseUsesSnakeCaseConversion() {
        let json = """
        {
            "access_token": "test_token",
            "refresh_token": "test_refresh",
            "expires_in": 3600,
            "user_id": 123,
            "account_id": "acc_456"
        }
        """.data(using: .utf8)!
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let token = try? decoder.decode(Token.self, from: json)
        XCTAssertNotNil(token)
        XCTAssertEqual(token?.accessToken, "test_token")
        XCTAssertEqual(token?.userId, 123)
    }

    func testLoginRequestEncoding() {
        let request = LoginRequest(phoneNumber: "13800138000", password: "test1234")
        let encoder = JSONEncoder()
        let data = try? encoder.encode(request)
        XCTAssertNotNil(data)
        let json = try? JSONSerialization.jsonObject(with: data!) as? [String: Any]
        XCTAssertEqual(json?["credential"] as? String, "13800138000")
        XCTAssertEqual(json?["password"] as? String, "test1234")
    }

    func testRefreshTokenRequestEncoding() {
        let request = RefreshTokenRequest(refreshToken: "refresh_val")
        let encoder = JSONEncoder()
        let data = try? encoder.encode(request)
        XCTAssertNotNil(data)
        let json = try? JSONSerialization.jsonObject(with: data!) as? [String: Any]
        XCTAssertEqual(json?["refresh_token"] as? String, "refresh_val")
    }
}
