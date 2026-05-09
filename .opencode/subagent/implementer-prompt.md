You are an expert software engineer tasked with implementing Task 6: Device Fingerprint Service (设备指纹服务) from the Account Center microservice implementation plan.

## Task Description
**Task 6: Device Fingerprint Service (设备指纹服务)**

**Files:**
- Create: `device-fingerprint-service/internal/service/device_service.go`
- Create: `device-fingerprint-service/internal/handler/device_handler.go`
- Create: `device-fingerprint-service/internal/model/device.go`
- Modify: `account-service/internal/service/user_service.go` (add trusted device check)
- Modify: `auth-service/internal/service/auth_service.go` (integrate device fingerprint check)
- Modify: `migrations/001_initial_schema.sql` (add device_fingerprints table if not already there)

**Steps to complete:**
1. Create device fingerprint model
2. Create device fingerprint service with risk assessment
3. Create device fingerprint handler
4. Integrate device fingerprint check in auth service
5. Update user service for trusted device management
6. Run tests
7. Commit

**Technical Details:**
- Use Go for backend service
- PostgreSQL for storing device fingerprints
- FingerprintJS concepts for device identification (though actual fingerprinting happens on frontend)
- Risk assessment based on device feature changes and geolocation
- Integration with auth service for login flow

**Requirements from Spec:**
- Device fingerprint generation via FingerprintJS (frontend) with backend storage and validation
- Trusted device mechanism: "最近一次强验证通过时间 + N 天", where N is 0-60 days, default 3 days
- Risk-aware trigger: 
  - Geolocation change detection (IP library) - force re-verification on drastic location change
  - Device fingerprint change detection - if key features change beyond threshold, force re-verification
- Store device features for risk assessment
- Trusted device bypasses second-factor authentication when within valid period
- Integration with login flow to check if device is trusted

**Best Practices:**
- Follow TDD: write failing tests first, then implement
- Proper error handling with meaningful error messages
- Input validation at handler level
- Repository pattern for data access
- Service layer for business logic
- Context cancellation support
- No hardcoded values (use constants/config)
- Follow Go conventions and formatting

**Deliverables:**
- All files listed above with complete implementation
- Working device fingerprint service with trust assessment
- Risk detection based on geolocation and device features
- Integration with authentication flow
- Clean, maintainable code following Go best practices