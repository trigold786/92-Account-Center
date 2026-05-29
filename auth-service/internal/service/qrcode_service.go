package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
)

type QRCodeService struct {
	rdb       *redis.Client
	jwtMgr    *jwt.JWTManager
	roleRepo  RoleRepository
	qrcodeTTL time.Duration
}

func NewQRCodeService(rdb *redis.Client, jwtMgr *jwt.JWTManager, roleRepo RoleRepository, qrcodeTTL time.Duration) *QRCodeService {
	if qrcodeTTL <= 0 {
		qrcodeTTL = 5 * time.Minute
	}
	return &QRCodeService{
		rdb:       rdb,
		jwtMgr:    jwtMgr,
		roleRepo:  roleRepo,
		qrcodeTTL: qrcodeTTL,
	}
}

type qrcodeData struct {
	CodeID string `json:"code_id"`
	Status string `json:"status"`
	UserID int64  `json:"user_id,omitempty"`
	Token  string `json:"token,omitempty"`
}

func (s *QRCodeService) Generate(ctx context.Context) (*model.QRCodeGenerateResponse, error) {
	codeID := uuid.New().String()
	data := qrcodeData{
		CodeID: codeID,
		Status: model.QRCodeStatusPending,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("qrcode:%s", codeID)
	if err := s.rdb.Set(ctx, key, jsonData, s.qrcodeTTL).Err(); err != nil {
		return nil, err
	}

	return &model.QRCodeGenerateResponse{
		CodeID:    codeID,
		ExpiresIn: int(s.qrcodeTTL.Seconds()),
	}, nil
}

func (s *QRCodeService) GetStatus(ctx context.Context, codeID string) (*model.QRCodeStatusResponse, error) {
	key := fmt.Sprintf("qrcode:%s", codeID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &model.QRCodeStatusResponse{
				CodeID: codeID,
				Status: model.QRCodeStatusExpired,
			}, nil
		}
		return nil, err
	}

	var data qrcodeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	resp := &model.QRCodeStatusResponse{
		CodeID: data.CodeID,
		Status: data.Status,
	}

	if data.Status == model.QRCodeStatusConfirmed && data.Token != "" {
		resp.Token = data.Token
	}

	return resp, nil
}

func (s *QRCodeService) Scan(ctx context.Context, codeID string, userID int64) error {
	key := fmt.Sprintf("qrcode:%s", codeID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("QR code expired or not found")
		}
		return err
	}

	var data qrcodeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return err
	}

	if data.Status != model.QRCodeStatusPending {
		return errors.New("QR code is not in pending status")
	}

	data.Status = model.QRCodeStatusScanned
	data.UserID = userID

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = s.qrcodeTTL
	}

	return s.rdb.Set(ctx, key, jsonData, ttl).Err()
}

func (s *QRCodeService) Confirm(ctx context.Context, codeID string, userID int64) (*model.QRCodeStatusResponse, error) {
	key := fmt.Sprintf("qrcode:%s", codeID)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("QR code expired or not found")
		}
		return nil, err
	}

	var data qrcodeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	if data.Status != model.QRCodeStatusScanned {
		return nil, errors.New("QR code must be scanned before confirmation")
	}

	if data.UserID != userID {
		return nil, errors.New("user ID mismatch")
	}

	accountID, roles, err := s.roleRepo.GetUserRolesByUserID(ctx, userID)
	if err != nil {
		slog.Default().Error("failed to get user roles for qrcode", "user_id", userID, "error", err.Error())
		accountID = fmt.Sprintf("%d", userID)
		roles = []string{}
	}

	tokenResp, err := s.jwtMgr.GenerateTokenPairWithDevice(userID, accountID, "", roles)
	if err != nil {
		return nil, err
	}

	tokenJSON, _ := json.Marshal(map[string]string{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
	})

	data.Status = model.QRCodeStatusConfirmed
	data.Token = string(tokenJSON)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		ttl = s.qrcodeTTL
	}

	if err := s.rdb.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return nil, err
	}

	return &model.QRCodeStatusResponse{
		CodeID: data.CodeID,
		Status: data.Status,
		Token:  data.Token,
	}, nil
}
