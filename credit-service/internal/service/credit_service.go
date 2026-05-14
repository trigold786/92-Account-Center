package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
	"github.com/trigold786/92-Account-Center/credit-service/pkg/crypto"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountFrozen       = errors.New("account is frozen")
	ErrAccountNotFound     = errors.New("account not found")
)

type CreditService interface {
	GetAccount(ctx context.Context, userID int64) (*model.AccountResponse, error)
	GetTransactions(ctx context.Context, userID int64, page, pageSize int) (*model.TransactionListResponse, error)
	EarnCredits(ctx context.Context, userID int64, amount float64, txnType, referenceID, details string) error
	ConsumeCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error
	RefundCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error
	CalculateDiscount(ctx context.Context, userID int64, price float64) (*model.CalculateDiscountResponse, error)
	VerifyIntegrity(ctx context.Context) (bool, error)
}

type creditService struct {
	creditRepo repository.CreditRepository
	db         *sql.DB
}

func NewCreditService(creditRepo repository.CreditRepository, db *sql.DB) CreditService {
	return &creditService{creditRepo: creditRepo, db: db}
}

func (s *creditService) getOrCreateAccount(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	account, err := s.creditRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account, err = s.creditRepo.CreateAccount(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return account, nil
}

func (s *creditService) GetAccount(ctx context.Context, userID int64) (*model.AccountResponse, error) {
	account, err := s.getOrCreateAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.AccountResponse{
		UserID:  account.UserID,
		Balance: account.Balance,
		Status:  account.Status,
	}, nil
}

func (s *creditService) GetTransactions(ctx context.Context, userID int64, page, pageSize int) (*model.TransactionListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	account, err := s.creditRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &model.TransactionListResponse{
			Transactions: []model.CreditTransaction{},
			Total:        0,
			Page:         page,
			PageSize:     pageSize,
		}, nil
	}

	transactions, total, err := s.creditRepo.GetTransactionsByAccountID(ctx, account.ID, page, pageSize)
	if err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []model.CreditTransaction{}
	}

	return &model.TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func (s *creditService) computeSM3Hash(txn *model.CreditTransaction, prevHash string) string {
	data := fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s",
		txn.ID, txn.CreditAccountID, txn.Type,
		strconv.FormatFloat(txn.Amount, 'f', 2, 64),
		txn.ReferenceID, txn.CreatedAt, prevHash)
	return crypto.SM3Hash(data)
}

func (s *creditService) EarnCredits(ctx context.Context, userID int64, amount float64, txnType, referenceID, details string) error {
	if referenceID != "" {
		existing, err := s.creditRepo.GetTransactionByReferenceID(ctx, referenceID)
		if err != nil {
			return err
		}
		if existing != nil {
			return nil
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := s.getOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	if account.Status == "FROZEN" {
		return ErrAccountFrozen
	}

	txn := &model.CreditTransaction{
		CreditAccountID: account.ID,
		Type:            txnType,
		Amount:          amount,
		ReferenceID:     referenceID,
		Details:         details,
		Status:          "COMPLETED",
	}

	lastTxn, err := s.creditRepo.GetLastTransaction(ctx, account.ID)
	if err != nil {
		return err
	}

	if err := s.creditRepo.CreateTransaction(ctx, txn); err != nil {
		return err
	}

	prevHash := ""
	if lastTxn != nil {
		prevHash = lastTxn.SM3Hash
	}
	txn.SM3Hash = s.computeSM3Hash(txn, prevHash)

	updateHashQuery := `UPDATE credit_transactions SET sm3_hash = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateHashQuery, txn.SM3Hash, txn.ID); err != nil {
		return err
	}

	if err := s.creditRepo.UpdateBalance(ctx, account.ID, amount); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *creditService) ConsumeCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error {
	existing, err := s.creditRepo.GetTransactionByReferenceID(ctx, referenceID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := s.creditRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrAccountNotFound
	}
	if account.Status == "FROZEN" {
		return ErrAccountFrozen
	}

	txn := &model.CreditTransaction{
		CreditAccountID: account.ID,
		Type:            "CONSUME_SUB",
		Amount:          amount,
		ReferenceID:     referenceID,
		Details:         details,
		Status:          "COMPLETED",
	}

	lastTxn, err := s.creditRepo.GetLastTransaction(ctx, account.ID)
	if err != nil {
		return err
	}

	if err := s.creditRepo.CreateTransaction(ctx, txn); err != nil {
		return err
	}

	prevHash := ""
	if lastTxn != nil {
		prevHash = lastTxn.SM3Hash
	}
	txn.SM3Hash = s.computeSM3Hash(txn, prevHash)

	updateHashQuery := `UPDATE credit_transactions SET sm3_hash = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateHashQuery, txn.SM3Hash, txn.ID); err != nil {
		return err
	}

	if err := s.creditRepo.UpdateBalance(ctx, account.ID, -amount); err != nil {
		return ErrInsufficientBalance
	}

	return tx.Commit()
}

func (s *creditService) RefundCredits(ctx context.Context, userID int64, amount float64, referenceID, details string) error {
	return s.EarnCredits(ctx, userID, amount, "REFUND_SUB", referenceID, details)
}

func (s *creditService) CalculateDiscount(ctx context.Context, userID int64, price float64) (*model.CalculateDiscountResponse, error) {
	account, err := s.creditRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &model.CalculateDiscountResponse{
			AvailableBalance: 0,
			MaxDiscount:      0,
			RemainingToPay:   price,
		}, nil
	}

	available := account.Balance
	maxDiscount := available
	if maxDiscount > price {
		maxDiscount = price
	}
	remaining := price - maxDiscount

	return &model.CalculateDiscountResponse{
		AvailableBalance: available,
		MaxDiscount:      maxDiscount,
		RemainingToPay:   remaining,
	}, nil
}

func (s *creditService) VerifyIntegrity(ctx context.Context) (bool, error) {
	transactions, err := s.creditRepo.GetAllTransactionsOrdered(ctx)
	if err != nil {
		return false, err
	}

	prevHash := ""
	for _, txn := range transactions {
		expected := fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s",
			txn.ID, txn.CreditAccountID, txn.Type,
			strconv.FormatFloat(txn.Amount, 'f', 2, 64),
			txn.ReferenceID, txn.CreatedAt, prevHash)
		computedHash := crypto.SM3Hash(expected)

		if computedHash != txn.SM3Hash {
			return false, nil
		}
		prevHash = txn.SM3Hash
	}

	return true, nil
}
