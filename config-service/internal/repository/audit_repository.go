package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type AuditRepository interface {
	ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error)
	GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error)
	CreateLog(ctx context.Context, log *model.AuditLog) error
}

type auditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.OperationType != "" {
		where = append(where, fmt.Sprintf("operation_type = $%d", argIdx))
		args = append(args, filter.OperationType)
		argIdx++
	}
	if filter.Operator != "" {
		where = append(where, fmt.Sprintf("operator = $%d", argIdx))
		args = append(args, filter.Operator)
		argIdx++
	}
	if filter.StartTime != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.StartTime)
		argIdx++
	}
	if filter.EndTime != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.EndTime)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs" + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	dataQuery := fmt.Sprintf(`SELECT id, operation_type, operation_object, operator, operator_ip,
		operation_result, operation_details, sm3_hash, created_at
		FROM audit_logs%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.OperationType, &l.OperationObject, &l.Operator,
			&l.OperatorIP, &l.OperationResult, &l.OperationDetails, &l.SM3Hash, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}

func (r *auditRepository) GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error) {
	l := &model.AuditLog{}
	query := `SELECT id, operation_type, operation_object, operator, operator_ip,
		operation_result, operation_details, sm3_hash, created_at FROM audit_logs WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&l.ID, &l.OperationType, &l.OperationObject,
		&l.Operator, &l.OperatorIP, &l.OperationResult, &l.OperationDetails, &l.SM3Hash, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return l, err
}

func (r *auditRepository) CreateLog(ctx context.Context, l *model.AuditLog) error {
	query := `INSERT INTO audit_logs (operation_type, operation_object, operator, operator_ip,
		operation_result, operation_details, sm3_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, l.OperationType, l.OperationObject, l.Operator,
		l.OperatorIP, l.OperationResult, l.OperationDetails, l.SM3Hash).Scan(&l.ID, &l.CreatedAt)
}

// Ensure time import is used
var _ = time.Now
