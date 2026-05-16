import XCTest
@testable import AccountCenter

final class TokenManagerTests: XCTestCase {
    var sut: TokenManager!

    override func setUp() {
        super.setUp()
        sut = TokenManager.shared
    }

    override func tearDown() {
        sut.clear()
        sut = nil
        super.tearDown()
    }

    func testSaveAndRetrieveToken() {
        let token = Token(
            accessToken: "access_123",
            refreshToken: "refresh_456",
            expiresIn: 3600,
            userId: 42,
            accountId: "acc_789"
        )
        sut.save(token: token)
        XCTAssertEqual(sut.getAccessToken(), "access_123")
        XCTAssertEqual(sut.getRefreshToken(), "refresh_456")
        XCTAssertEqual(sut.getUserId(), 42)
        XCTAssertEqual(sut.getAccountId(), "acc_789")
    }

    func testClearToken() {
        let token = Token(
            accessToken: "access_123",
            refreshToken: "refresh_456",
            expiresIn: 3600,
            userId: 42,
            accountId: "acc_789"
        )
        sut.save(token: token)
        sut.clear()
        XCTAssertNil(sut.getAccessToken())
        XCTAssertNil(sut.getRefreshToken())
    }
}
