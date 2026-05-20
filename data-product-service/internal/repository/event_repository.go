package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) BatchInsert(ctx context.Context, events []model.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO events (event_type, user_id, session_id, device_id, properties, timestamp)
		 VALUES ($1,$2,$3,$4,$5,$6)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range events {
		propsJSON, _ := json.Marshal(e.Properties)
		ts := e.Timestamp
		if ts.IsZero() {
			ts = time.Now()
		}
		if _, err := stmt.ExecContext(ctx, e.EventType, e.UserID, e.SessionID, e.DeviceID, propsJSON, ts); err != nil {
			return err
		}
	}
	return tx.Commit()
}
