package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
)

type mockCreditRepo struct {
	mock.Mock
}

func (m *mockCreditRepo) CreateAccount(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CreditAccount), args.Error(1)
}

func (m *mockCreditRepo) GetAccountByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CreditAccount), args.Error(1)
}

func (m *mockCreditRepo) GetAccountByID(ctx context.Context, id int64) (*model.CreditAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CreditAccount), args.Error(1)
}

func (m *mockCreditRepo) UpdateBalance(ctx context.Context, id int64, delta float64) error {
	args := m.Called(ctx, id, delta)
	return args.Error(0)
}

func (m *mockCreditRepo) CreateTransaction(ctx context.Context, txn *model.CreditTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *mockCreditRepo) GetTransactionsByAccountID(ctx context.Context, accountID int64, page, pageSize int) ([]model.CreditTransaction, int, error) {
	args := m.Called(ctx, accountID, page, pageSize)
	return args.Get(0).([]model.CreditTransaction), args.Int(1), args.Error(2)
}

func (m *mockCreditRepo) GetLastTransaction(ctx context.Context, accountID int64) (*model.CreditTransaction, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CreditTransaction), args.Error(1)
}

func (m *mockCreditRepo) GetTransactionByReferenceID(ctx context.Context, referenceID string) (*model.CreditTransaction, error) {
	args := m.Called(ctx, referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CreditTransaction), args.Error(1)
}

func (m *mockCreditRepo) UpdateTransactionStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockCreditRepo) GetAllTransactionsOrdered(ctx context.Context) ([]model.CreditTransaction, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.CreditTransaction), args.Error(1)
}

func (m *mockCreditRepo) GetRebateConfig(ctx context.Context, subscriptionCount int) (*model.RebateConfig, error) {
	args := m.Called(ctx, subscriptionCount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RebateConfig), args.Error(1)
}

type mockReferralRepo struct {
	mock.Mock
}

func (m *mockReferralRepo) Create(ctx context.Context, referrerID, refereeID int64) (*model.ReferralRelation, error) {
	args := m.Called(ctx, referrerID, refereeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReferralRelation), args.Error(1)
}

func (m *mockReferralRepo) GetByRefereeID(ctx context.Context, refereeID int64) (*model.ReferralRelation, error) {
	args := m.Called(ctx, refereeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReferralRelation), args.Error(1)
}

func (m *mockReferralRepo) GetByReferrerID(ctx context.Context, referrerID int64) ([]model.ReferralRelation, error) {
	args := m.Called(ctx, referrerID)
	return args.Get(0).([]model.ReferralRelation), args.Error(1)
}

func (m *mockReferralRepo) IncrementSubscriptionCount(ctx context.Context, refereeID int64) error {
	args := m.Called(ctx, refereeID)
	return args.Error(0)
}

func (m *mockReferralRepo) GetReferralSummary(ctx context.Context, referrerID int64) (*model.ReferralSummary, error) {
	args := m.Called(ctx, referrerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReferralSummary), args.Error(1)
}

type mockCreditService struct {
	mock.Mock
}

func (m *mockCreditService) GetAccount(ctx context.Context, userID int64) (*model.AccountResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccountResponse), args.Error(1)
}

func (m *mockCreditService) GetTransactions(ctx context.Context, userID int64, page, pageSize int) (*model.TransactionListResponse, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TransactionListResponse), args.Error(1)
}

func (m *mockCreditService) EarnCredits(ctx context.Context, userID int64, amount float64, txnType, referenceID, details string) error {
	args := m.Called(ctx, userID, amount, txnType, referenceID, details)
	return args.Error(0)
}

func (m *mockCreditService) ConsumeCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error {
	args := m.Called(ctx, userID, amount, referenceID, details)
	return args.Error(0)
}

func (m *mockCreditService) RefundCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error {
	args := m.Called(ctx, userID, amount, referenceID, details)
	return args.Error(0)
}

func (m *mockCreditService) CalculateDiscount(ctx context.Context, userID int64, price float64) (*model.CalculateDiscountResponse, error) {
	args := m.Called(ctx, userID, price)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CalculateDiscountResponse), args.Error(1)
}

func (m *mockCreditService) VerifyIntegrity(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
