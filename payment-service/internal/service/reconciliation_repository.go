package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type sqlReconciliationReportRepository struct {
	db *sql.DB
}

func NewSQLReconciliationReportRepository(db *sql.DB) ReconciliationReportRepository {
	return &sqlReconciliationReportRepository{db: db}
}

func (r *sqlReconciliationReportRepository) Save(ctx context.Context, report *ReconciliationReport) error {
	mismatches, err := json.Marshal(report.MismatchOrders)
	if err != nil {
		return err
	}
	createdAt := report.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	query := `
		INSERT INTO reconciliation_reports (id, provider_name, report_date, total_orders, matched_orders, mismatch_orders, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			provider_name = EXCLUDED.provider_name,
			report_date = EXCLUDED.report_date,
			total_orders = EXCLUDED.total_orders,
			matched_orders = EXCLUDED.matched_orders,
			mismatch_orders = EXCLUDED.mismatch_orders,
			status = EXCLUDED.status`
	_, err = r.db.ExecContext(ctx, query,
		report.ID,
		report.ProviderName,
		report.Date,
		report.TotalOrders,
		report.MatchedOrders,
		string(mismatches),
		report.Status,
		createdAt,
	)
	return err
}

func (r *sqlReconciliationReportRepository) GetByID(ctx context.Context, id string) (*ReconciliationReport, error) {
	var report ReconciliationReport
	var mismatchJSON []byte
	query := `SELECT id, provider_name, report_date::text, total_orders, matched_orders, mismatch_orders, status, created_at FROM reconciliation_reports WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.ProviderName,
		&report.Date,
		&report.TotalOrders,
		&report.MatchedOrders,
		&mismatchJSON,
		&report.Status,
		&report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(mismatchJSON) > 0 {
		if err := json.Unmarshal(mismatchJSON, &report.MismatchOrders); err != nil {
			return nil, err
		}
	}
	if report.MismatchOrders == nil {
		report.MismatchOrders = []MismatchOrder{}
	}
	return &report, nil
}
