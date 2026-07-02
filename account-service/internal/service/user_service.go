package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/argon2"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/account-service/pkg/crypto"
	"github.com/trigold786/92-Account-Center/account-service/pkg/sms"
)

var (
	ErrPhoneAlreadyRegistered = errors.New("phone number already registered")
	ErrAccountIDTaken         = errors.New("account ID already taken")
	ErrMustAgreeToTerms       = errors.New("must agree to terms and conditions")
	ErrInvalidPhoneNumber     = errors.New("invalid phone number format")
	ErrInvalidAccountID       = errors.New("invalid account ID format")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrSamePassword           = errors.New("new password must be different from current")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)

var (
	phoneRegex     = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	accountIDRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{5,19}$`)
)

type UserService interface {
	Register(ctx context.Context, phoneNumber, accountID, password string, agreeToTerms bool) (*model.User, error)
	ValidatePassword(password string) error
	ChangePassword(ctx context.Context, userID string, req *model.ChangePasswordRequest) error
	SendPasswordVerificationCode(ctx context.Context, contactType, contactValue string) error
	RequestAccountDeletion(ctx context.Context, userID string, req *model.DeletionRequest) (*model.DeletionResponse, error)
	CancelAccountDeletion(ctx context.Context, userID string) (*model.DeletionResponse, error)
	GetDeletionStatus(ctx context.Context, userID string) (*model.Deletion, error)
	GetProfile(ctx context.Context, userID int64) (*model.UserProfile, error)
	UpdateEmail(ctx context.Context, userID int64, email string) error
	UpdatePhone(ctx context.Context, userID int64, phone string) error
}

type userService struct {
	repo      repository.UserRepository
	smsClient *sms.Client
	cfg       *svcconfig.AccountConfig
}

func NewUserService(repo repository.UserRepository, smsClient *sms.Client, cfg *svcconfig.AccountConfig) UserService {
	return &userService{repo: repo, smsClient: smsClient, cfg: cfg}
}

func (s *userService) Register(ctx context.Context, phoneNumber, accountID, password string, agreeToTerms bool) (*model.User, error) {
	if !phoneRegex.MatchString(phoneNumber) {
		return nil, ErrInvalidPhoneNumber
	}

	if !accountIDRegex.MatchString(accountID) {
		if len(accountID) > 0 && accountID[0] >= '0' && accountID[0] <= '9' {
			return nil, fmt.Errorf("%w: cannot start with a number", ErrInvalidAccountID)
		}
		return nil, ErrInvalidAccountID
	}

	if err := s.validatePasswordStrength(password); err != nil {
		return nil, err
	}

	if !agreeToTerms {
		return nil, ErrMustAgreeToTerms
	}

	exists, err := s.repo.ExistsByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyRegistered
	}

	exists, err = s.repo.ExistsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountIDTaken
	}

	passwordHash, err := hashPasswordArgon2id(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &model.User{
		PhoneNumber:  phoneNumber,
		AccountID:    accountID,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ValidatePassword(password string) error {
	return s.validatePasswordStrength(password)
}

func (s *userService) ChangePassword(ctx context.Context, userID string, req *model.ChangePasswordRequest) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if err := s.validatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	if req.NewPassword != req.ConfirmPassword {
		return errors.New("passwords do not match")
	}

	passwordHash, err := hashPasswordArgon2id(req.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	return s.repo.Update(ctx, user)
}

func (s *userService) SendPasswordVerificationCode(ctx context.Context, contactType, contactValue string) error {
	if s.smsClient == nil {
		return nil
	}
	if contactType == "phone" {
		return s.smsClient.SendCode(contactValue)
	}
	return nil
}

func (s *userService) RequestAccountDeletion(ctx context.Context, userID string, req *model.DeletionRequest) (*model.DeletionResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.DeletionDeletedAt != nil {
		return nil, errors.New("account is already deleted")
	}

	if user.DeletionRequestedAt != nil &&
		(user.DeletionCancelledAt == nil || user.DeletionRequestedAt.After(*user.DeletionCancelledAt)) &&
		(user.DeletionDeletedAt == nil || user.DeletionRequestedAt.After(*user.DeletionDeletedAt)) {
		return nil, errors.New("deletion already requested for this account")
	}

	switch req.VerificationType {
	case "sms_code":
		if req.VerificationCode == "" {
			return nil, errors.New("verification code required")
		}
		if s.smsClient == nil {
			return nil, errors.New("SMS verification service is not configured")
		}
		if user.PhoneNumber == "" {
			return nil, errors.New("no phone number on file for verification")
		}
		valid, err := s.smsClient.VerifyCode(user.PhoneNumber, req.VerificationCode)
		if err != nil {
			return nil, fmt.Errorf("verification failed: %w", err)
		}
		if !valid {
			return nil, errors.New("invalid or expired verification code")
		}
	case "email_otp":
		return nil, errors.New("email OTP verification is not yet supported; please use SMS code")
	default:
		return nil, errors.New("invalid verification type")
	}

	now := time.Now()
	freezePeriod := time.Duration(s.cfg.DeletionFreezeDays) * 24 * time.Hour
	user.DeletionRequestedAt = &now
	expiresAt := now.Add(freezePeriod)
	user.DeletionExpiresAt = &expiresAt
	user.DeletionCancelledAt = nil
	user.DeletionDeletedAt = nil
	user.UpdatedAt = now

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &model.DeletionResponse{
		Message: "Account deletion requested successfully. Account will be permanently deleted after the freeze period.",
	}, nil
}

func (s *userService) CancelAccountDeletion(ctx context.Context, userID string) (*model.DeletionResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.DeletionRequestedAt == nil {
		return nil, errors.New("no deletion request found for this account")
	}

	if user.DeletionDeletedAt != nil && user.DeletionRequestedAt.After(*user.DeletionDeletedAt) {
		return nil, errors.New("cannot cancel deletion as account is already deleted")
	}

	if user.DeletionCancelledAt != nil && user.DeletionRequestedAt.After(*user.DeletionCancelledAt) {
		return nil, errors.New("deletion request already cancelled")
	}

	now := time.Now()
	user.DeletionCancelledAt = &now
	user.UpdatedAt = now

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &model.DeletionResponse{
		Message: "Account deletion request cancelled successfully.",
	}, nil
}

func (s *userService) GetDeletionStatus(ctx context.Context, userID string) (*model.Deletion, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	return &model.Deletion{
		UserID:      user.ID,
		RequestedAt: user.DeletionRequestedAt,
		ExpiresAt:   user.DeletionExpiresAt,
		CancelledAt: user.DeletionCancelledAt,
		DeletedAt:   user.DeletionDeletedAt,
	}, nil
}

func (s *userService) validatePasswordStrength(password string) error {
	if len(password) < 8 || len(password) > 20 {
		return fmt.Errorf("password must be between 8 and 20 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func generateSalt() string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(fmt.Sprintf("generate salt: %v", err))
	}
	return hex.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	data := []byte(salt + password)
	return crypto.SM3Hash(data)
}

func hashPasswordArgon2id(password string) (string, error) {
	const (
		argon2idTime    = 3
		argon2idMemory  = 64 * 1024
		argon2idThreads = 2
		argon2idKeyLen  = 32
		argon2idSaltLen = 16
	)
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate argon2id salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2idMemory,
		argon2idTime,
		argon2idThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, storedHash string) bool {
	if strings.HasPrefix(storedHash, "$argon2id$") {
		return verifyPasswordArgon2id(password, storedHash)
	}
	parts := splitStoredHash(storedHash)
	if len(parts) != 2 {
		return false
	}
	salt, hash := parts[0], parts[1]
	computedHash := hashPassword(password, salt)
	return subtleCompare(hash, computedHash)
}

func verifyPasswordArgon2id(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	var memory, timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expectedHash)))
	return subtleCompare(string(expectedHash), string(computedHash))
}

func splitStoredHash(stored string) []string {
	for i := len(stored) - 1; i >= 0; i-- {
		if stored[i] == '$' {
			return []string{stored[:i], stored[i+1:]}
		}
	}
	return nil
}

func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func ParseUserID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func IsValidPhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidAccountID(accountID string) bool {
	return accountIDRegex.MatchString(accountID)
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*model.UserProfile, error) {
	user, err := s.repo.GetByID(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &model.UserProfile{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		AccountID:   user.AccountID,
		Email:       *user.Email,
		MFAEnabled:  user.MFAEnabled,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (s *userService) UpdateEmail(ctx context.Context, userID int64, email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return s.repo.UpdateEmail(ctx, userID, email)
}

func (s *userService) UpdatePhone(ctx context.Context, userID int64, phone string) error {
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhoneNumber
	}
	exists, err := s.repo.ExistsByPhoneNumber(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return ErrPhoneAlreadyRegistered
	}
	return s.repo.UpdatePhone(ctx, userID, phone)
}
