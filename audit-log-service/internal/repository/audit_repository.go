package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/model"
)

type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	CreateBatch(ctx context.Context, logs []*model.AuditLog) error
	GetByLogID(ctx context.Context, logID string) (*model.AuditLog, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.AuditLog, error)
	GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*model.AuditLog, error)
	DeleteOlderThan(ctx context.Context, cutoffTime time.Time) (int64, error)
}

type auditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (log_id, user_id, event_time, action_type, target_resource, source_ip, result, details, sm3_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`
	return r.db.QueryRowContext(ctx, query,
		log.LogID,
		log.UserID,
		log.EventTime,
		log.ActionType,
		log.TargetResource,
		log.SourceIP,
		log.Result,
		log.Details,
		log.SM3Hash,
		log.CreatedAt,
	).Scan(&log.CreatedAt)
}

func (r *auditRepository) CreateBatch(ctx context.Context, logs []*model.AuditLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO audit_logs (log_id, user_id, event_time, action_type, target_resource, source_ip, result, details, sm3_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, log := range logs {
		_, err := stmt.ExecContext(ctx,
			log.LogID,
			log.UserID,
			log.EventTime,
			log.ActionType,
			log.TargetResource,
			log.SourceIP,
			log.Result,
			log.Details,
			log.SM3Hash,
			log.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *auditRepository) GetByLogID(ctx context.Context, logID string) (*model.AuditLog, error) {
	log := &model.AuditLog{}
	var details []byte
	query := `SELECT log_id, user_id, event_time, action_type, target_resource, source_ip, result, details, sm3_hash, created_at FROM audit_logs WHERE log_id = $1`
	err := r.db.QueryRowContext(ctx, query, logID).Scan(
		&log.LogID,
		&log.UserID,
		&log.EventTime,
		&log.ActionType,
		&log.TargetResource,
		&log.SourceIP,
		&log.Result,
		&details,
		&log.SM3Hash,
		&log.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	log.Details = json.RawMessage(details)
	return log, nil
}

func (r *auditRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.AuditLog, error) {
	query := `
		SELECT log_id, user_id, event_time, action_type, target_resource, source_ip, result, details, sm3_hash, created_at 
		FROM audit_logs 
		WHERE user_id = $1 
		ORDER BY event_time DESC 
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *auditRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	query := `
		SELECT log_id, user_id, event_time, action_type, target_resource, source_ip, result, details, sm3_hash, created_at 
		FROM audit_logs 
		WHERE event_time >= $1 AND event_time <= $2 
		ORDER BY event_time DESC 
		LIMIT $3 OFFSET $4`
	rows, err := r.db.QueryContext(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *auditRepository) DeleteOlderThan(ctx context.Context, cutoffTime time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE event_time < $1`
	result, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *auditRepository) scanLogs(rows *sql.Rows) ([]*model.AuditLog, error) {
	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var details []byte
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.EventTime,
			&log.ActionType,
			&log.TargetResource,
			&log.SourceIP,
			&log.Result,
			&details,
			&log.SM3Hash,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		log.Details = json.RawMessage(details)
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
