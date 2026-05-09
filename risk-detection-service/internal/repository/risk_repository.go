package repository

import (
	"context"
	"database/sql"
	"time"

	"risk-detection-service/internal/model"
)

type RiskRepository struct {
	db *sql.DB
}

func NewRiskRepository(db *sql.DB) *RiskRepository {
	return &RiskRepository{db: db}
}

func (r *RiskRepository) Create(ctx context.Context, event *model.RiskEvent) error {
	query := `
		INSERT INTO risk_events (risk_event_id, user_id, event_type, risk_score, risk_level, details, ip_address, location, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		event.RiskEventID,
		event.UserID,
		event.EventType,
		event.RiskScore,
		event.RiskLevel,
		event.Details,
		event.IPAddress,
		event.Location,
		event.CreatedAt,
	)

	return err
}

func (r *RiskRepository) GetByUserID(ctx context.Context, userID string, start, end time.Time, limit int) ([]*model.RiskEvent, error) {
	query := `
		SELECT risk_event_id, user_id, event_type, risk_score, risk_level, details, ip_address, location, created_at
		FROM risk_events
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.RiskEvent
	for rows.Next() {
		event := &model.RiskEvent{}
		err := rows.Scan(
			&event.RiskEventID,
			&event.UserID,
			&event.EventType,
			&event.RiskScore,
			&event.RiskLevel,
			&event.Details,
			&event.IPAddress,
			&event.Location,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *RiskRepository) GetByEventID(ctx context.Context, eventID string) (*model.RiskEvent, error) {
	query := `
		SELECT risk_event_id, user_id, event_type, risk_score, risk_level, details, ip_address, location, created_at
		FROM risk_events
		WHERE risk_event_id = $1
	`

	event := &model.RiskEvent{}
	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.RiskEventID,
		&event.UserID,
		&event.EventType,
		&event.RiskScore,
		&event.RiskLevel,
		&event.Details,
		&event.IPAddress,
		&event.Location,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *RiskRepository) GetLastEventByUserID(ctx context.Context, userID string) (*model.RiskEvent, error) {
	query := `
		SELECT risk_event_id, user_id, event_type, risk_score, risk_level, details, ip_address, location, created_at
		FROM risk_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	event := &model.RiskEvent{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&event.RiskEventID,
		&event.UserID,
		&event.EventType,
		&event.RiskScore,
		&event.RiskLevel,
		&event.Details,
		&event.IPAddress,
		&event.Location,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *RiskRepository) CountEventsByUserIDSince(ctx context.Context, userID string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM risk_events
		WHERE user_id = $1 AND created_at >= $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, since).Scan(&count)
	return count, err
}