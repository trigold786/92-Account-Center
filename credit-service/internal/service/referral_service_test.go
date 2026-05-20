package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/svcconfig"
)

func TestReferralService_BindReferral_InvalidCode(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	err := svc.BindReferral(context.Background(), "!!!invalid-base64!!!", "5")
	assert.ErrorIs(t, err, ErrInvalidReferralCode)
}

func TestReferralService_BindReferral_InvalidRefereeID(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(1)
	err := svc.BindReferral(context.Background(), code, "not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid referee_id")
}

func TestReferralService_BindReferral_SelfReferral(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(1)
	err := svc.BindReferral(context.Background(), code, "1")
	assert.ErrorIs(t, err, ErrSelfReferral)
}

func TestReferralService_BindReferral_AlreadyReferred(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(10)
	repo.On("GetByRefereeID", mock.Anything, int64(5)).
		Return(&model.ReferralRelation{ID: 1, ReferrerID: 10, RefereeID: 5}, nil)

	err := svc.BindReferral(context.Background(), code, "5")
	assert.ErrorIs(t, err, ErrAlreadyReferred)
	repo.AssertExpectations(t)
}

func TestReferralService_BindReferral_GetByRefereeError(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(10)
	repo.On("GetByRefereeID", mock.Anything, int64(5)).
		Return(nil, errors.New("db error"))

	err := svc.BindReferral(context.Background(), code, "5")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestReferralService_BindReferral_Success(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(10)
	repo.On("GetByRefereeID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("Create", mock.Anything, int64(10), int64(5)).
		Return(&model.ReferralRelation{ID: 1, ReferrerID: 10, RefereeID: 5}, nil)

	err := svc.BindReferral(context.Background(), code, "5")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestReferralService_BindReferral_CreateError(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	code := encodeReferralCode(10)
	repo.On("GetByRefereeID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("Create", mock.Anything, int64(10), int64(5)).
		Return(nil, errors.New("unique violation"))

	err := svc.BindReferral(context.Background(), code, "5")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestReferralService_GenerateLink_Success(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := &svcconfig.CreditConfig{
		ReferralLinkTemplate: "https://app.test.com/ref?code=%s",
	}
	svc := NewReferralService(repo, cfg)

	resp, err := svc.GenerateLink(context.Background(), 42)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ReferralCode)
	assert.Contains(t, resp.ReferralLink, resp.ReferralCode)
	assert.Contains(t, resp.ReferralLink, "https://app.test.com/ref?code=")
}

func TestReferralService_GenerateLink_CodeIsDecodable(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	resp, err := svc.GenerateLink(context.Background(), 123)
	assert.NoError(t, err)

	decodedID, err := decodeReferralCode(resp.ReferralCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), decodedID)
}

func TestReferralService_GetSummary_Delegates(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	expected := &model.ReferralSummary{
		TotalReferees:  5,
		TotalEarned:    250.0,
		ActiveReferees: 3,
	}
	repo.On("GetReferralSummary", mock.Anything, int64(1)).Return(expected, nil)

	summary, err := svc.GetSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 5, summary.TotalReferees)
	assert.Equal(t, 250.0, summary.TotalEarned)
	assert.Equal(t, 3, summary.ActiveReferees)
	repo.AssertExpectations(t)
}

func TestReferralService_GetSummary_Error(t *testing.T) {
	repo := new(mockReferralRepo)
	cfg := defaultCreditConfig()
	svc := NewReferralService(repo, cfg)

	repo.On("GetReferralSummary", mock.Anything, int64(1)).
		Return(nil, errors.New("db error"))

	summary, err := svc.GetSummary(context.Background(), 1)
	assert.Nil(t, summary)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestReferralService_EncodeDecodeReferralCode_RoundTrip(t *testing.T) {
	ids := []int64{1, 100, 999999, 0}
	for _, id := range ids {
		code := encodeReferralCode(id)
		decoded, err := decodeReferralCode(code)
		assert.NoError(t, err)
		assert.Equal(t, id, decoded, "roundtrip failed for id=%d", id)
	}
}

func TestReferralService_DecodeReferralCode_EmptyString(t *testing.T) {
	_, err := decodeReferralCode("")
	assert.ErrorIs(t, err, ErrInvalidReferralCode)
}

func TestReferralService_DecodeReferralCode_TamperedCode(t *testing.T) {
	code := encodeReferralCode(42)
	tampered := code + "x"
	_, err := decodeReferralCode(tampered)
	assert.ErrorIs(t, err, ErrInvalidReferralCode)
}

func TestReferralService_DecodeReferralCode_NoColon(t *testing.T) {
	_, err := decodeReferralCode("aW52YWxpZG5vY29sb24")
	assert.ErrorIs(t, err, ErrInvalidReferralCode)
}

func TestReferralService_DecodeReferralCode_InvalidUserID(t *testing.T) {
	_, err := decodeReferralCode("YWJjOmRlZg")
	assert.ErrorIs(t, err, ErrInvalidReferralCode)
}
