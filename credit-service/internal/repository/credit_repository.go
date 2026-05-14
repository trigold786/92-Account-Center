package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
)

type CreditRepository interface {
	CreateAccount(ctx context.Context, userID int64) (*model.CreditAccount, error)
	GetAccountByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error)
	GetAccountByID(ctx context.Context, id int64) (*model.CreditAccount, error)
	UpdateBalance(ctx context.Context, id int64, delta float64) error
	CreateTransaction(ctx context.Context, txn *model.CreditTransaction) error
	GetTransactionsByAccountID(ctx context.Context, accountID int64, page, pageSize int) ([]model.CreditTransaction, int, error)
	GetLastTransaction(ctx context.Context, accountID int64) (*model.CreditTransaction, error)
	GetTransactionByReferenceID(ctx context.Context, referenceID string) (*model.CreditTransaction, error)
	UpdateTransactionStatus(ctx context.Context, id int64, status string) error
	GetAllTransactionsOrdered(ctx context.Context) ([]model.CreditTransaction, error)
}

type creditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) CreditRepository {
	return &creditRepository{db: db}
}

func (r *creditRepository) CreateAccount(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	account := &model.CreditAccount{}
	query := `
		INSERT INTO credit_accounts (user_id, balance, status, created_at, updated_at)
		VALUES ($1, 0, 'ACTIVE', NOW(), NOW())
		RETURNING id, user_id, balance, status, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Status,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *creditRepository) GetAccountByUserID(ctx context.Context, userID int64) (*model.CreditAccount, error) {
	account := &model.CreditAccount{}
	query := `SELECT id, user_id, balance, status, created_at, updated_at FROM credit_accounts WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Status,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return account, nil
}

func (r *creditRepository) GetAccountByID(ctx context.Context, id int64) (*model.CreditAccount, error) {
	account := &model.CreditAccount{}
	query := `SELECT id, user_id, balance, status, created_at, updated_at FROM credit_accounts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.Status,
		&account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return account, nil
}

func (r *creditRepository) UpdateBalance(ctx context.Context, id int64, delta float64) error {
	query := `UPDATE credit_accounts SET balance = balance + $1, updated_at = NOW() WHERE id = $2 AND balance + $1 >= 0`
	result, err := r.db.ExecContext(ctx, query, delta, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient balance or account not found")
	}
	return nil
}

func (r *creditRepository) CreateTransaction(ctx context.Context, txn *model.CreditTransaction) error {
	query := `
		INSERT INTO credit_transactions (credit_account_id, type, amount, reference_id, details, sm3_hash, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		txn.CreditAccountID, txn.Type, txn.Amount, txn.ReferenceID, txn.Details, txn.SM3Hash, txn.Status,
	).Scan(&txn.ID, &txn.CreatedAt)
}

func (r *creditRepository) GetTransactionsByAccountID(ctx context.Context, accountID int64, page, pageSize int) ([]model.CreditTransaction, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM credit_transactions WHERE credit_account_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, accountID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, credit_account_id, type, amount, reference_id, details, sm3_hash, status, created_at
		FROM credit_transactions WHERE credit_account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, accountID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []model.CreditTransaction
	for rows.Next() {
		var txn model.CreditTransaction
		if err := rows.Scan(
			&txn.ID, &txn.CreditAccountID, &txn.Type, &txn.Amount,
			&txn.ReferenceID, &txn.Details, &txn.SM3Hash, &txn.Status, &txn.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, txn)
	}
	return transactions, total, rows.Err()
}

func (r *creditRepository) GetLastTransaction(ctx context.Context, accountID int64) (*model.CreditTransaction, error) {
	txn := &model.CreditTransaction{}
	query := `SELECT id, credit_account_id, type, amount, reference_id, details, sm3_hash, status, created_at
		FROM credit_transactions WHERE credit_account_id = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&txn.ID, &txn.CreditAccountID, &txn.Type, &txn.Amount,
		&txn.ReferenceID, &txn.Details, &txn.SM3Hash, &txn.Status, &txn.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return txn, nil
}

func (r *creditRepository) GetTransactionByReferenceID(ctx context.Context, referenceID string) (*model.CreditTransaction, error) {
	txn := &model.CreditTransaction{}
	query := `SELECT id, credit_account_id, type, amount, reference_id, details, sm3_hash, status, created_at
		FROM credit_transactions WHERE reference_id = $1`
	err := r.db.QueryRowContext(ctx, query, referenceID).Scan(
		&txn.ID, &txn.CreditAccountID, &txn.Type, &txn.Amount,
		&txn.ReferenceID, &txn.Details, &txn.SM3Hash, &txn.Status, &txn.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return txn, nil
}

func (r *creditRepository) UpdateTransactionStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE credit_transactions SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *creditRepository) GetAllTransactionsOrdered(ctx context.Context) ([]model.CreditTransaction, error) {
	query := `SELECT id, credit_account_id, type, amount, reference_id, details, sm3_hash, status, created_at
		FROM credit_transactions ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.CreditTransaction
	for rows.Next() {
		var txn model.CreditTransaction
		if err := rows.Scan(
			&txn.ID, &txn.CreditAccountID, &txn.Type, &txn.Amount,
			&txn.ReferenceID, &txn.Details, &txn.SM3Hash, &txn.Status, &txn.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, txn)
	}
	return transactions, rows.Err()
}
