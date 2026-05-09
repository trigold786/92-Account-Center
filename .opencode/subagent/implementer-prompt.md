You are an expert software engineer tasked with implementing Task 1: User Registration (手机号注册) from the Account Center microservice implementation plan.

## Task Description
**Task 1: User Registration (手机号注册)**

**Files:**
- Create: `account-service/internal/model/user.go`
- Create: `account-service/internal/repository/user_repository.go`
- Create: `account-service/internal/service/user_service.go`
- Create: `account-service/internal/handler/register_handler.go`
- Create: `account-service/pkg/crypto/sm4.go`
- Create: `account-service/pkg/crypto/sm3.go`
- Create: `migrations/001_initial_schema.sql`
- Modify: `api-gateway/internal/router/router.go`

**Steps to complete:**
1. Write database migration for users table
2. Create user model
3. Create SM3 hash utility
4. Create user repository
5. Create user service with password hashing and validation
6. Create register handler
7. Add routes to API gateway
8. Run tests
9. Commit

**Technical Details:**
- Use Go/Gin for backend services
- PostgreSQL for database
- Redis for caching (though not used in this task specifically)
- SM3 for hash utilities
- bcrypt for password hashing
- JWT for tokens (though more relevant to auth service)
- Gin framework for HTTP handlers
- govalidator or similar for input validation (using struct tags)

**Requirements from Spec:**
- User registration via phone number + SMS verification code
- Account ID must be 6-20 chars, letters/numbers/underscore, not starting with number
- Password must meet security policy (8-20 chars, upper/lower/number/special char)
- Must agree to terms
- Phone number verification with 5-minute expiry
- Global uniqueness check for phone number and account ID
- Password hashed with bcrypt + salt
- Generate JWT access/refresh tokens on successful registration

**Best Practices:**
- Follow TDD: write failing tests first, then implement
- Proper error handling with meaningful error messages
- Input validation at handler level
- Repository pattern for data access
- Service layer for business logic
- Proper logging (though not explicitly required in this task)
- Return appropriate HTTP status codes
- Context cancellation support
- No hardcoded values (use constants/config)
- Follow Go conventions and formatting

**Deliverables:**
- All files listed above with complete implementation
- Working user registration endpoint
- Proper error handling and validation
- Database schema with appropriate indexes
- Clean, maintainable code following Go best practices