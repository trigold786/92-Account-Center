package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/sunxi/92-Account-Center/account-service/internal/model"
	"github.com/sunxi/92-Account-Center/account-service/internal/repository"
	"github.com/sunxi/92-Account-Center/account-service/pkg/cache"
	"github.com/sunxi/92-Account-Center/account-service/pkg/crypto"
)

// UserService defines the interface for user service.
type UserService interface {
	Register(ctx context.Context, phoneNumber, accountID, password string, agreeToTerms bool) (*model.User, error)
	ValidatePassword(password string) error
	ChangePassword(ctx context.Context, userID string, req *model.ChangePasswordRequest) error
	SendPasswordVerificationCode(ctx context.Context, contactType, contactValue string) error
}

// userService implements UserService.
type userService struct {
	repo       repository.UserRepository
	sessionCache cache.SessionCache
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository, sessionCache cache.SessionCache) UserService {
	return &userService{repo: repo, sessionCache: sessionCache}
}

// Register creates a new user with the given details.
func (s *userService) Register(ctx context.Context, phoneNumber, accountID, password string, agreeToTerms bool) (*model.User, error) {
	// Validate input
	if err := validatePhoneNumber(phoneNumber); err != nil {
		return nil, err
	}
	if err := validateAccountID(accountID); err != nil {
		return nil, err
	}
	if err := s.ValidatePassword(password); err != nil {
		return nil, err
	}
	if !agreeToTerms {
		return nil, errors.New("must agree to terms and conditions")
	}

	// Check if phone number or account ID already exists
	exists, err := s.repo.ExistsByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("phone number already registered")
	}

	exists, err = s.repo.ExistsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("account ID already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		PhoneNumber:  phoneNumber,
		AccountID:    accountID,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ValidatePassword checks if password meets security policy.
func (s *userService) ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 20 {
		return errors.New("password must be between 8 and 20 characters")
	}
	
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= '!' && char <= '/': // Special characters
			hasSpecial = true
		case char >= ':' && char <= '@': // Special characters
			hasSpecial = true
		case char >= '[' && char <= '`': // Special characters
			hasSpecial = true
		case char >= '{' && char <= '~': // Special characters
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}

// validatePhoneNumber checks if phone number is valid.
func validatePhoneNumber(phoneNumber string) error {
	// Simple validation for Chinese phone numbers (11 digits starting with 1)
	matched, err := regexp.MatchString(`^1[3-9]\d{9}$`, phoneNumber)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid phone number format")
	}
	return nil
}

// validateAccountID checks if account ID is valid.
func validateAccountID(accountID string) error {
	// Account ID must be 6-20 chars, letters/numbers/underscore, not starting with number
	matched, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{5,19}$`, accountID)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("account ID must be 6-20 characters, letters/numbers/underscore, and cannot start with a number")
	}
	return nil
}

// ChangePassword changes a user's password with verification.
func (s *userService) ChangePassword(ctx context.Context, userID string, req *model.ChangePasswordRequest) error {
	// Get user by ID
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if user.IsLocked {
		return errors.New("account is locked")
	}

	// Verify identity based on verification type
	switch req.VerificationType {
	case "password":
		if req.CurrentPassword == "" {
			return errors.New("current password required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword+user.PasswordSalt)); err != nil {
			return errors.New("invalid current password")
		}
	case "sms_code", "email_otp":
		if req.VerificationCode == "" {
			return errors.New("verification code required")
		}
		// In a real implementation, we would verify the code against stored values
		// For now, we'll just check if it's not empty as a placeholder
		if req.VerificationCode == "" {
			return errors.New("invalid verification code")
		}
	default:
		return errors.New("invalid verification type")
	}

	// Check if new passwords match
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("new password and confirmation do not match")
	}

	// Validate new password strength
	if err := s.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	// Generate new salt and hash new password
	newSalt := uuid.New().String()
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword+newSalt), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.PasswordHash = string(newHash)
	user.PasswordSalt = newSalt
	user.UpdatedAt = time.Now()

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}

	// Invalidate all user sessions
	if err := s.sessionCache.InvalidateUserSessions(ctx, userID); err != nil {
		return err
	}

	return nil
}

// SendPasswordVerificationCode sends a verification code for password change.
func (s *userService) SendPasswordVerificationCode(ctx context.Context, contactType, contactValue string) error {
	// In a real implementation, this would send an SMS or email with a verification code
	// For now, we'll just return nil as a placeholder
	return nil
}