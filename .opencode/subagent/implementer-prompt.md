You are an expert software engineer tasked with implementing Task 5: SMS/Email Service with Circuit Breaker (多云短信服务) from the Account Center microservice implementation plan.

## Task Description
**Task 5: SMS/Email Service with Circuit Breaker (多云短信服务)**

**Files:**
- Create: `sms-email-service/internal/provider/aliyun.go`
- Create: `sms-email-service/internal/provider/tencent.go`
- Create: `sms-email-service/internal/provider/chinatelecom.go`
- Create: `sms-email-service/internal/service/sms_service.go`
- Create: `sms-email-service/pkg/circuitbreaker/circuitbreaker.go`

**Steps to complete:**
1. Create circuit breaker utility
2. Create SMS provider interface and implementations
3. Create SMS service with provider failover
4. Create SMS handler
5. Run tests
6. Commit

**Technical Details:**
- Use Go for backend service
- Redis for rate limiting and verification code storage
- Circuit breaker pattern for provider failover
- Support for multiple SMS providers (Aliyun, Tencent, Chinatelecom)
- Rate limiting (120s interval, 10 per day per phone number)
- Verification code generation and validation
- Integration with existing services via message queue (Kafka)

**Requirements from Spec:**
- Multi-cloud SMS service with 阿里云, 腾讯云, 天翼云
- Strict frequency control: 120s interval between sends, 10 per day limit per phone number
- Verification code validity: 5 minutes
- Automatic failover when primary channel error rate >15% in 5 minutes
- Alerting to operations personnel when failover occurs
- Support for both SMS verification codes and email OTP/Magic Link
- Integration with Kafka for asynchronous communication

**Best Practices:**
- Follow TDD: write failing tests first, then implement
- Proper error handling with meaningful error messages
- Circuit breaker pattern for resilience
- Rate limiting to prevent abuse
- Context cancellation support
- No hardcoded values (use constants/config)
- Follow Go conventions and formatting
- Provider abstraction for easy extension

**Deliverables:**
- All files listed above with complete implementation
- Working SMS service with provider failover
- Rate limiting implementation
- Verification code generation and storage
- Circuit breaker functionality
- Clean, maintainable code following Go best practices