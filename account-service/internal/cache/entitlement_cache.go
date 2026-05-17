package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

const (
	entitlementKeyPrefix = "entitlement:"
	consumeQuotaScript   = `
local key = KEYS[1]
local field = ARGV[1]
local amount = tonumber(ARGV[2])
local data = redis.call('HGET', key, field)
if data then
    local obj = cjson.decode(data)
    if obj.total - obj.used >= amount then
        obj.used = obj.used + amount
        redis.call('HSET', key, field, cjson.encode(obj))
        return 1
    else
        return 0
    end
else
    return -1
end
`
)

type EntitlementCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewEntitlementCache(rdb *redis.Client, ttl time.Duration) *EntitlementCache {
	return &EntitlementCache{rdb: rdb, ttl: ttl}
}

func (c *EntitlementCache) cacheKey(userID int64) string {
	return entitlementKeyPrefix + strconv.FormatInt(userID, 10)
}

func (c *EntitlementCache) WarmCache(ctx context.Context, entitlements []model.Entitlement) error {
	if len(entitlements) == 0 {
		return nil
	}
	key := c.cacheKey(entitlements[0].UserID)
	pipe := c.rdb.Pipeline()
	for _, e := range entitlements {
		data, err := json.Marshal(model.EntitlementQuota{
			Total: e.TotalQuota,
			Used:  e.UsedQuota,
		})
		if err != nil {
			return err
		}
		pipe.HSet(ctx, key, e.FeatureCode, string(data))
	}
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *EntitlementCache) GetQuota(ctx context.Context, userID int64, featureCode string) (*model.EntitlementQuota, error) {
	key := c.cacheKey(userID)
	data, err := c.rdb.HGet(ctx, key, featureCode).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var quota model.EntitlementQuota
	if err := json.Unmarshal([]byte(data), &quota); err != nil {
		return nil, err
	}
	return &quota, nil
}

func (c *EntitlementCache) ConsumeQuota(ctx context.Context, userID int64, featureCode string, amount int) (bool, error) {
	key := c.cacheKey(userID)
	result, err := c.rdb.Eval(ctx, consumeQuotaScript, []string{key}, featureCode, amount).Int()
	if err != nil {
		return false, err
	}
	switch result {
	case 1:
		return true, nil
	case 0:
		return false, fmt.Errorf("insufficient quota")
	default:
		return false, fmt.Errorf("entitlement not found in cache")
	}
}

func (c *EntitlementCache) GrantQuota(ctx context.Context, userID int64, featureCode string, total int) error {
	key := c.cacheKey(userID)
	data, err := json.Marshal(model.EntitlementQuota{
		Total: total,
		Used:  0,
	})
	if err != nil {
		return err
	}
	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, featureCode, string(data))
	pipe.Expire(ctx, key, c.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *EntitlementCache) InvalidateCache(ctx context.Context, userID int64) error {
	return c.rdb.Del(ctx, c.cacheKey(userID)).Err()
}
