package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
	"github.com/trigold786/92-Account-Center/notification-service/internal/provider"
)

type PushService struct {
	rdb      *redis.Client
	registry *provider.PushProviderRegistry
}

func NewPushService(rdb *redis.Client, registry *provider.PushProviderRegistry) *PushService {
	return &PushService{rdb: rdb, registry: registry}
}

func (s *PushService) SendPush(ctx context.Context, req *model.PushRequest) (*model.PushResponse, error) {
	var err error

	switch req.Platform {
	case model.PushPlatformIOS:
		err = s.sendViaProvider(ctx, "apns", req.DeviceToken, req.Title, req.Body, req.Data)
	case model.PushPlatformAndroid:
		err = s.sendViaProvider(ctx, "fcm", req.DeviceToken, req.Title, req.Body, req.Data)
	case model.PushPlatformHuawei:
		err = s.sendViaProvider(ctx, "hms", req.DeviceToken, req.Title, req.Body, req.Data)
	case model.PushPlatformXiaomi:
		log.Printf("[Xiaomi] Sending notification to user %s: %s", req.UserID, req.Title)
	case model.PushPlatformOppo:
		log.Printf("[Oppo] Sending notification to user %s: %s", req.UserID, req.Title)
	case model.PushPlatformVivo:
		log.Printf("[Vivo] Sending notification to user %s: %s", req.UserID, req.Title)
	case model.PushPlatformWeb:
		log.Printf("[WebPush] Sending notification to user %s: %s", req.UserID, req.Title)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", req.Platform)
	}

	if err != nil {
		log.Printf("Failed to send push notification: %v", err)
		return &model.PushResponse{
			Code:    500,
			Message: "推送发送失败",
		}, err
	}

	s.recordPushHistory(ctx, req)

	return &model.PushResponse{
		Code:    200,
		Message: "推送发送成功",
	}, nil
}

func (s *PushService) SendPushNotification(ctx context.Context, req *model.PushSendRequest) (*model.PushResponse, error) {
	devices, err := s.GetUserDeviceTokens(ctx, req.UserID)
	if err != nil {
		return &model.PushResponse{Code: 500, Message: "获取设备列表失败"}, err
	}

	if len(devices) == 0 {
		return &model.PushResponse{Code: 404, Message: "未找到已注册设备"}, nil
	}

	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	var lastErr error
	for _, d := range devices {
		if req.Platform != "" && d.Platform != req.Platform {
			continue
		}

		providerName := s.platformToProvider(d.Platform)
		if err := s.sendViaProvider(ctx, providerName, d.DeviceToken, req.Title, req.Body, data); err != nil {
			log.Printf("Failed to send push to %s via %s: %v", d.DeviceToken, providerName, err)
			lastErr = err
		}
	}

	if lastErr != nil {
		return &model.PushResponse{Code: 500, Message: "部分推送发送失败"}, lastErr
	}

	return &model.PushResponse{Code: 200, Message: "推送发送成功"}, nil
}

func (s *PushService) RegisterDevice(ctx context.Context, device *model.PushDevice) error {
	key := fmt.Sprintf("push:device:%s:%s", device.UserID, device.DeviceToken)
	data, _ := json.Marshal(device)
	return s.rdb.Set(ctx, key, data, 0).Err()
}

func (s *PushService) RegisterDeviceToken(ctx context.Context, req *model.DeviceTokenRequest) error {
	key := fmt.Sprintf("push_tokens:%s", req.UserID)
	member := fmt.Sprintf("%s:%s", req.Platform, req.DeviceToken)
	if err := s.rdb.SAdd(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to register device token: %w", err)
	}

	deviceKey := fmt.Sprintf("push:device:%s:%s", req.UserID, req.DeviceToken)
	device := &model.PushDevice{
		ID:          fmt.Sprintf("%s-%s", req.UserID, req.DeviceToken[:8]),
		UserID:      req.UserID,
		DeviceToken: req.DeviceToken,
		Platform:    model.PushPlatform(req.Platform),
		IsActive:    true,
		CreatedAt:   &[]time.Time{time.Now()}[0],
	}
	data, _ := json.Marshal(device)
	return s.rdb.Set(ctx, deviceKey, data, 0).Err()
}

func (s *PushService) UnregisterDeviceToken(ctx context.Context, userID, deviceToken string) error {
	pattern := fmt.Sprintf("push_tokens:%s", userID)
	members, err := s.rdb.SMembers(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get device tokens: %w", err)
	}

	for _, m := range members {
		parts := strings.SplitN(m, ":", 2)
		if len(parts) == 2 && parts[1] == deviceToken {
			if err := s.rdb.SRem(ctx, pattern, m).Err(); err != nil {
				return fmt.Errorf("failed to unregister device token: %w", err)
			}
			break
		}
	}

	deviceKey := fmt.Sprintf("push:device:%s:%s", userID, deviceToken)
	s.rdb.Del(ctx, deviceKey)

	return nil
}

func (s *PushService) GetUserDevices(ctx context.Context, userID string) ([]*model.PushDevice, error) {
	pattern := fmt.Sprintf("push:device:%s:*", userID)
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	devices := make([]*model.PushDevice, 0)
	for _, key := range keys {
		data, err := s.rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		var device model.PushDevice
		if err := json.Unmarshal([]byte(data), &device); err == nil {
			devices = append(devices, &device)
		}
	}

	return devices, nil
}

func (s *PushService) GetUserDeviceTokens(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
	key := fmt.Sprintf("push_tokens:%s", userID)
	members, err := s.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	tokens := make([]*model.DeviceToken, 0, len(members))
	for _, m := range members {
		parts := strings.SplitN(m, ":", 2)
		if len(parts) != 2 {
			continue
		}
		tokens = append(tokens, &model.DeviceToken{
			UserID:       userID,
			Platform:     parts[0],
			DeviceToken:  parts[1],
			RegisteredAt: "",
		})
	}

	return tokens, nil
}

func (s *PushService) sendViaProvider(ctx context.Context, providerName, deviceToken, title, body string, data map[string]interface{}) error {
	if s.registry == nil {
		log.Printf("[%s] registry not configured, skipping send", providerName)
		return nil
	}

	p, ok := s.registry.Get(providerName)
	if !ok {
		log.Printf("[%s] provider not registered, skipping send", providerName)
		return nil
	}

	pushData := make(map[string]string)
	for k, v := range data {
		pushData[k] = fmt.Sprintf("%v", v)
	}

	resp, err := p.Send(ctx, &provider.PushRequest{
		DeviceToken: deviceToken,
		Title:       title,
		Body:        body,
		Data:        pushData,
	})
	if err != nil {
		return fmt.Errorf("%s send failed: %w", providerName, err)
	}
	if !resp.Success {
		return fmt.Errorf("%s send failed: %s", providerName, resp.Error)
	}

	return nil
}

func (s *PushService) platformToProvider(platform string) string {
	switch platform {
	case "ios":
		return "apns"
	case "android":
		return "fcm"
	case "huawei":
		return "hms"
	default:
		return "fcm"
	}
}

func (s *PushService) recordPushHistory(ctx context.Context, req *model.PushRequest) {
	if s.rdb == nil {
		return
	}
	key := fmt.Sprintf("push:history:%s", req.UserID)
	data, _ := json.Marshal(req)
	s.rdb.LPush(ctx, key, data)
	s.rdb.LTrim(ctx, key, 0, 99)
}
