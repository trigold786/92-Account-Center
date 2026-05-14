package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type DeviceRepository interface {
	Save(ctx context.Context, fp *model.DeviceFingerprint) error
	GetByFingerprintID(ctx context.Context, userID uint64, fingerprintID string) (*model.DeviceFingerprint, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error)
	Update(ctx context.Context, fp *model.DeviceFingerprint) error
	Delete(ctx context.Context, id uint64) error
}

type deviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Save(ctx context.Context, fp *model.DeviceFingerprint) error {
	existing, _ := r.GetByFingerprintID(ctx, fp.UserID, fp.FingerprintID)
	if existing != nil {
		return r.Update(ctx, fp)
	}

	query := `INSERT INTO device_fingerprints (user_id, fingerprint_id, user_agent, ip_address, country, city, latitude, longitude, features, is_trusted, last_used_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		fp.UserID, fp.FingerprintID, fp.UserAgent, fp.IPAddress,
		fp.Country, fp.City, fp.Latitude, fp.Longitude,
		fp.Features, fp.IsTrusted, fp.LastUsedAt, fp.CreatedAt, fp.UpdatedAt,
	).Scan(&fp.ID)
}

func (r *deviceRepository) GetByFingerprintID(ctx context.Context, userID uint64, fingerprintID string) (*model.DeviceFingerprint, error) {
	fp := &model.DeviceFingerprint{}
	query := `SELECT id, user_id, fingerprint_id, user_agent, ip_address, country, city, latitude, longitude, features, is_trusted, last_used_at, created_at, updated_at
		FROM device_fingerprints WHERE user_id = $1 AND fingerprint_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, fingerprintID).Scan(
		&fp.ID, &fp.UserID, &fp.FingerprintID, &fp.UserAgent, &fp.IPAddress,
		&fp.Country, &fp.City, &fp.Latitude, &fp.Longitude,
		&fp.Features, &fp.IsTrusted, &fp.LastUsedAt, &fp.CreatedAt, &fp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func (r *deviceRepository) GetByUserID(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error) {
	query := `SELECT id, user_id, fingerprint_id, user_agent, ip_address, country, city, latitude, longitude, features, is_trusted, last_used_at, created_at, updated_at
		FROM device_fingerprints WHERE user_id = $1 ORDER BY last_used_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*model.DeviceFingerprint
	for rows.Next() {
		fp := &model.DeviceFingerprint{}
		if err := rows.Scan(
			&fp.ID, &fp.UserID, &fp.FingerprintID, &fp.UserAgent, &fp.IPAddress,
			&fp.Country, &fp.City, &fp.Latitude, &fp.Longitude,
			&fp.Features, &fp.IsTrusted, &fp.LastUsedAt, &fp.CreatedAt, &fp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, fp)
	}
	return devices, rows.Err()
}

func (r *deviceRepository) Update(ctx context.Context, fp *model.DeviceFingerprint) error {
	query := `UPDATE device_fingerprints SET user_agent = $1, ip_address = $2, country = $3, city = $4, latitude = $5, longitude = $6, features = $7, is_trusted = $8, last_used_at = $9, updated_at = $10
		WHERE user_id = $11 AND fingerprint_id = $12`
	result, err := r.db.ExecContext(ctx, query,
		fp.UserAgent, fp.IPAddress, fp.Country, fp.City,
		fp.Latitude, fp.Longitude, fp.Features, fp.IsTrusted,
		fp.LastUsedAt, fp.UpdatedAt, fp.UserID, fp.FingerprintID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

func (r *deviceRepository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM device_fingerprints WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}
