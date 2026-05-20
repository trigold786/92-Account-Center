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

func defaultCreditConfig() *svcconfig.CreditConfig {
	return &svcconfig.CreditConfig{
		DefaultPageSize:           20,
		DefaultRebateRate:         0.10,
		ReferralLinkTemplate:      "https://app.example.com/referral?code=%s",
		SubscriptionStreamKey:     "subscription:paid",
		SubscriptionConsumerGroup: "credit-rebate-group",
		SubscriptionConsumerID:    "credit-worker-1",
	}
}

func ptrAccount(a model.CreditAccount) *model.CreditAccount {
	return &a
}

func TestCreditService_GetAccount_Success(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1, Balance: 100.5, Status: "ACTIVE"}), nil)

	resp, err := svc.GetAccount(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.UserID)
	assert.Equal(t, 100.5, resp.Balance)
	assert.Equal(t, "ACTIVE", resp.Status)
	repo.AssertExpectations(t)
}

func TestCreditService_GetAccount_CreatesWhenNotFound(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(2)).Return(nil, nil)
	repo.On("CreateAccount", mock.Anything, int64(2)).
		Return(ptrAccount(model.CreditAccount{ID: 20, UserID: 2, Balance: 0, Status: "ACTIVE"}), nil)

	resp, err := svc.GetAccount(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.UserID)
	assert.Equal(t, 0.0, resp.Balance)
	repo.AssertExpectations(t)
}

func TestCreditService_GetAccount_GetAccountError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	dbErr := errors.New("db connection lost")
	repo.On("GetAccountByUserID", mock.Anything, int64(1)).Return(nil, dbErr)

	resp, err := svc.GetAccount(context.Background(), 1)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, dbErr)
	repo.AssertExpectations(t)
}

func TestCreditService_GetAccount_CreateAccountError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	createErr := errors.New("unique violation")
	repo.On("GetAccountByUserID", mock.Anything, int64(3)).Return(nil, nil)
	repo.On("CreateAccount", mock.Anything, int64(3)).Return(nil, createErr)

	resp, err := svc.GetAccount(context.Background(), 3)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, createErr)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_DefaultPagination(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1}), nil)
	repo.On("GetTransactionsByAccountID", mock.Anything, int64(10), 1, 20).
		Return([]model.CreditTransaction{}, 0, nil)

	resp, err := svc.GetTransactions(context.Background(), 1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Equal(t, []model.CreditTransaction{}, resp.Transactions)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_CustomPagination(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txns := []model.CreditTransaction{
		{ID: 1, CreditAccountID: 10, Type: "EARN", Amount: 50},
		{ID: 2, CreditAccountID: 10, Type: "CONSUME", Amount: 20},
	}

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1}), nil)
	repo.On("GetTransactionsByAccountID", mock.Anything, int64(10), 2, 5).
		Return(txns, 2, nil)

	resp, err := svc.GetTransactions(context.Background(), 1, 2, 5)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 5, resp.PageSize)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Transactions, 2)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_NegativePageDefaultsTo1(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1}), nil)
	repo.On("GetTransactionsByAccountID", mock.Anything, int64(10), 1, 20).
		Return([]model.CreditTransaction{}, 0, nil)

	resp, err := svc.GetTransactions(context.Background(), 1, -1, -1)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_NilAccount_ReturnsEmpty(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(99)).Return(nil, nil)

	resp, err := svc.GetTransactions(context.Background(), 99, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Transactions)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_GetAccountError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(nil, errors.New("db error"))

	resp, err := svc.GetTransactions(context.Background(), 1, 1, 10)
	assert.Nil(t, resp)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_GetTransactionsError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1}), nil)
	repo.On("GetTransactionsByAccountID", mock.Anything, int64(10), 1, 10).
		Return([]model.CreditTransaction(nil), 0, errors.New("query error"))

	resp, err := svc.GetTransactions(context.Background(), 1, 1, 10)
	assert.Nil(t, resp)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_GetTransactions_NilTransactionsReturnsEmptySlice(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1}), nil)
	repo.On("GetTransactionsByAccountID", mock.Anything, int64(10), 1, 20).
		Return([]model.CreditTransaction(nil), 0, nil)

	resp, err := svc.GetTransactions(context.Background(), 1, 0, 0)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Transactions)
	assert.Empty(t, resp.Transactions)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_NilAccount(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).Return(nil, nil)

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, resp.AvailableBalance)
	assert.Equal(t, 0.0, resp.MaxDiscount)
	assert.Equal(t, 100.0, resp.RemainingToPay)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_BalanceLessThanPrice(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1, Balance: 30.0, Status: "ACTIVE"}), nil)

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 30.0, resp.AvailableBalance)
	assert.Equal(t, 30.0, resp.MaxDiscount)
	assert.Equal(t, 70.0, resp.RemainingToPay)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_BalanceGreaterThanPrice(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1, Balance: 200.0, Status: "ACTIVE"}), nil)

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 200.0, resp.AvailableBalance)
	assert.Equal(t, 100.0, resp.MaxDiscount)
	assert.Equal(t, 0.0, resp.RemainingToPay)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_BalanceEqualsPrice(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1, Balance: 100.0, Status: "ACTIVE"}), nil)

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, resp.AvailableBalance)
	assert.Equal(t, 100.0, resp.MaxDiscount)
	assert.Equal(t, 0.0, resp.RemainingToPay)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_ZeroBalance(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(ptrAccount(model.CreditAccount{ID: 10, UserID: 1, Balance: 0, Status: "ACTIVE"}), nil)

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, resp.AvailableBalance)
	assert.Equal(t, 0.0, resp.MaxDiscount)
	assert.Equal(t, 100.0, resp.RemainingToPay)
	repo.AssertExpectations(t)
}

func TestCreditService_CalculateDiscount_RepoError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAccountByUserID", mock.Anything, int64(1)).
		Return(nil, errors.New("db error"))

	resp, err := svc.CalculateDiscount(context.Background(), 1, 100.0)
	assert.Nil(t, resp)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_VerifyIntegrity_NoTransactions(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAllTransactionsOrdered", mock.Anything).
		Return([]model.CreditTransaction{}, nil)

	ok, err := svc.VerifyIntegrity(context.Background())
	assert.NoError(t, err)
	assert.True(t, ok)
	repo.AssertExpectations(t)
}

func TestCreditService_VerifyIntegrity_ValidChain(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txns := buildValidTransactionChain(t, 3)
	repo.On("GetAllTransactionsOrdered", mock.Anything).Return(txns, nil)

	ok, err := svc.VerifyIntegrity(context.Background())
	assert.NoError(t, err)
	assert.True(t, ok)
	repo.AssertExpectations(t)
}

func TestCreditService_VerifyIntegrity_InvalidChain_TamperedAmount(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txns := buildValidTransactionChain(t, 3)
	txns[1].Amount = 999.99

	repo.On("GetAllTransactionsOrdered", mock.Anything).Return(txns, nil)

	ok, err := svc.VerifyIntegrity(context.Background())
	assert.NoError(t, err)
	assert.False(t, ok)
	repo.AssertExpectations(t)
}

func TestCreditService_VerifyIntegrity_InvalidChain_TamperedHash(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txns := buildValidTransactionChain(t, 2)
	txns[0].SM3Hash = "deadbeef"

	repo.On("GetAllTransactionsOrdered", mock.Anything).Return(txns, nil)

	ok, err := svc.VerifyIntegrity(context.Background())
	assert.NoError(t, err)
	assert.False(t, ok)
	repo.AssertExpectations(t)
}

func TestCreditService_VerifyIntegrity_RepoError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetAllTransactionsOrdered", mock.Anything).
		Return([]model.CreditTransaction(nil), errors.New("db error"))

	ok, err := svc.VerifyIntegrity(context.Background())
	assert.Error(t, err)
	assert.False(t, ok)
	repo.AssertExpectations(t)
}

func TestCreditService_EarnCredits_DuplicateReference_ReturnsNil(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	existing := &model.CreditTransaction{ID: 1, ReferenceID: "ref:123"}
	repo.On("GetTransactionByReferenceID", mock.Anything, "ref:123").
		Return(existing, nil)

	err := svc.EarnCredits(context.Background(), 1, 50.0, "EARN", "ref:123", "{}")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_EarnCredits_GetRefError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetTransactionByReferenceID", mock.Anything, "ref:err").
		Return(nil, errors.New("db error"))

	err := svc.EarnCredits(context.Background(), 1, 50.0, "EARN", "ref:err", "{}")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_EarnCredits_EmptyReference_SkipsDedup(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	// With empty referenceID, the dedup check is skipped (line 118-126 of credit_service.go).
	// But we still can't proceed past BeginTx since db is nil, so we expect a nil pointer panic
	// or error. This test verifies the skip of the dedup path only.

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic due to nil db.BeginTx")
	}()

	_ = svc.EarnCredits(context.Background(), 1, 50.0, "EARN", "", "{}")
}

func TestCreditService_ConsumeCredits_DuplicateReference_ReturnsNil(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	existing := &model.CreditTransaction{ID: 1, ReferenceID: "order:456"}
	repo.On("GetTransactionByReferenceID", mock.Anything, "order:456").
		Return(existing, nil)

	err := svc.ConsumeCredits(context.Background(), 1, 30.0, "order:456", "{}")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_ConsumeCredits_GetRefError(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	repo.On("GetTransactionByReferenceID", mock.Anything, "order:err").
		Return(nil, errors.New("db error"))

	err := svc.ConsumeCredits(context.Background(), 1, 30.0, "order:err", "{}")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreditService_computeSM3Hash_Deterministic(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txn := &model.CreditTransaction{
		ID:              1,
		CreditAccountID: 10,
		Type:            "EARN",
		Amount:          50.0,
		ReferenceID:     "ref:1",
		CreatedAt:       "2025-01-01T00:00:00Z",
	}
	prevHash := "abc123"

	hash1 := svc.computeSM3Hash(txn, prevHash)
	hash2 := svc.computeSM3Hash(txn, prevHash)
	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)
}

func TestCreditService_computeSM3Hash_DifferentInputs(t *testing.T) {
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txn := &model.CreditTransaction{
		ID:              1,
		CreditAccountID: 10,
		Type:            "EARN",
		Amount:          50.0,
		ReferenceID:     "ref:1",
		CreatedAt:       "2025-01-01T00:00:00Z",
	}

	hash1 := svc.computeSM3Hash(txn, "")
	hash2 := svc.computeSM3Hash(txn, "prev")
	assert.NotEqual(t, hash1, hash2)
}

func buildValidTransactionChain(t *testing.T, count int) []model.CreditTransaction {
	t.Helper()
	repo := new(mockCreditRepo)
	cfg := defaultCreditConfig()
	svc := NewCreditService(repo, nil, cfg).(*creditService)

	txns := make([]model.CreditTransaction, count)
	prevHash := ""
	for i := 0; i < count; i++ {
		txn := model.CreditTransaction{
			ID:              int64(i + 1),
			CreditAccountID: 10,
			Type:            "EARN",
			Amount:          float64((i + 1) * 10),
			ReferenceID:     "",
			CreatedAt:       "2025-01-01T00:00:00Z",
		}
		txn.SM3Hash = svc.computeSM3Hash(&txn, prevHash)
		prevHash = txn.SM3Hash
		txns[i] = txn
	}
	return txns
}
