// SendPasswordVerificationCode sends a verification code for password change.
func (s *userService) SendPasswordVerificationCode(ctx context.Context, contactType, contactValue string) error {
	// In a real implementation, this would send an SMS or email with a verification code
	// For now, we'll just return nil as a placeholder
	return nil
}

// RequestAccountDeletion requests deletion of a user account with verification.
func (s *userService) RequestAccountDeletion(ctx context.Context, userID string, req *model.DeletionRequest) (*model.DeletionResponse, error) {
	// Validate user ID
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Get user by ID
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Check if account is already deleted
	if user.DeletionDeletedAt != nil {
		return nil, errors.New("account is already deleted")
	}

	// Check if deletion is already requested and not cancelled/expired
	if user.DeletionRequestedAt != nil && 
	   (user.DeletionCancelledAt == nil || user.DeletionRequestedAt.After(*user.DeletionCancelledAt)) &&
	   (user.DeletionDeletedAt == nil || user.DeletionRequestedAt.After(*user.DeletionDeletedAt)) {
		return nil, errors.New("deletion already requested for this account")
	}

	// Verify identity based on verification type
	switch req.VerificationType {
	case "sms_code", "email_otp":
		if req.VerificationCode == "" {
			return nil, errors.New("verification code required")
		}
		// In a real implementation, we would verify the code against stored values
		// For now, we'll just check if it's not empty as a placeholder
		if req.VerificationCode == "" {
			return nil, errors.New("invalid verification code")
		}
	default:
		return nil, errors.New("invalid verification type")
	}

	// Set deletion timestamps
	now := time.Now()
	user.DeletionRequestedAt = &now
	user.DeletionExpiresAt = &now.Add(model.FreezePeriod)
	user.DeletionCancelledAt = nil
	user.DeletionDeletedAt = nil
	user.UpdatedAt = now

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Invalidate all user sessions
	if err := s.sessionCache.InvalidateUserSessions(ctx, userID); err != nil {
		return nil, err
	}

	return &model.DeletionResponse{
		Message: "Account deletion requested successfully. Account will be permanently deleted after the freeze period.",
	}, nil
}

// CancelAccountDeletion cancels a pending account deletion request.
func (s *userService) CancelAccountDeletion(ctx context.Context, userID string) (*model.DeletionResponse, error) {
	// Validate user ID
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Get user by ID
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Check if deletion was requested
	if user.DeletionRequestedAt == nil {
		return nil, errors.New("no deletion request found for this account")
	}

	// Check if already deleted
	if user.DeletionDeletedAt != nil && user.DeletionRequestedAt.After(*user.DeletionDeletedAt) {
		return nil, errors.New("cannot cancel deletion as account is already deleted")
	}

	// Check if already cancelled
	if user.DeletionCancelledAt != nil && user.DeletionRequestedAt.After(*user.DeletionCancelledAt) {
		return nil, errors.New("deletion request already cancelled")
	}

	// Set cancellation timestamp
	now := time.Now()
	user.DeletionCancelledAt = &now
	user.UpdatedAt = now

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &model.DeletionResponse{
		Message: "Account deletion request cancelled successfully.",
	}, nil
}

// GetDeletionStatus retrieves the deletion status of a user account.
func (s *userService) GetDeletionStatus(ctx context.Context, userID string) (*model.Deletion, error) {
	// Validate user ID
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Get user by ID
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Return deletion status
	return &model.Deletion{
		UserID:    user.ID,
		RequestedAt: user.DeletionRequestedAt,
		ExpiresAt:   user.DeletionExpiresAt,
		CancelledAt: user.DeletionCancelledAt,
		DeletedAt:   user.DeletionDeletedAt,
	}, nil
}

// IsTrustedDevice checks if a device is trusted for the user based on device fingerprint service
func (s *userService) IsTrustedDevice(ctx context.Context, userID string, fingerprintID string) (bool, error) {
	// Validate inputs
	if userID == "" {
		return false, errors.New("user ID is required")
	}
	if fingerprintID == "" {
		return false, errors.New("fingerprint ID is required")
	}

	// Convert userID to uint64
	var uid uint64
	parsedID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}
	uid = parsedID

	// Call device fingerprint service to check if device is trusted
	isTrusted, err := s.deviceFingerprintService.IsTrusted(ctx, uid, fingerprintID)
	if err != nil {
		return false, fmt.Errorf("failed to check device trust status: %w", err)
	}

	return isTrusted, nil
}