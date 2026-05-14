package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type PushService struct {
	rdb *redis.Client
}

func NewPushService(rdb *redis.Client) *PushService {
	return &PushService{rdb: rdb}
}

func (s *PushService) SendPush(ctx context.Context, req *model.PushRequest) (*model.PushResponse, error) {
	var err error

	switch req.Platform {
	case model.PushPlatformIOS:
		err = s.sendAPNs(ctx, req)
	case model.PushPlatformAndroid:
		err = s.sendFCM(ctx, req)
	case model.PushPlatformXiaomi:
		err = s.sendXiaomi(ctx, req)
	case model.PushPlatformHuawei:
		err = s.sendHuawei(ctx, req)
	case model.PushPlatformOppo:
		err = s.sendOppo(ctx, req)
	case model.PushPlatformVivo:
		err = s.sendVivo(ctx, req)
	case model.PushPlatformWeb:
		err = s.sendWebPush(ctx, req)
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

func (s *PushService) RegisterDevice(ctx context.Context, device *model.PushDevice) error {
	key := fmt.Sprintf("push:device:%s:%s", device.UserID, device.DeviceToken)
	data, _ := json.Marshal(device)
	return s.rdb.Set(ctx, key, data, 0).Err()
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

func (s *PushService) sendAPNs(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending APNs notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendFCM(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending FCM notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendXiaomi(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending Xiaomi notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendHuawei(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending Huawei notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendOppo(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending Oppo notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendVivo(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending Vivo notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) sendWebPush(ctx context.Context, req *model.PushRequest) error {
	log.Printf("Sending Web Push notification to user %s: %s", req.UserID, req.Title)
	return nil
}

func (s *PushService) recordPushHistory(ctx context.Context, req *model.PushRequest) {
	key := fmt.Sprintf("push:history:%s", req.UserID)
	data, _ := json.Marshal(req)
	s.rdb.LPush(ctx, key, data)
	s.rdb.LTrim(ctx, key, 0, 99)
}
