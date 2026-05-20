package saga

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type StateStore interface {
	Save(ctx context.Context, saga *Saga) error
	Load(ctx context.Context, id string) (*Saga, error)
}

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(client *redis.Client, ttl time.Duration) *RedisStore {
	return &RedisStore{client: client, ttl: ttl}
}

func (s *RedisStore) Save(ctx context.Context, saga *Saga) error {
	data, err := json.Marshal(saga)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "saga:"+saga.ID, data, s.ttl).Err()
}

func (s *RedisStore) Load(ctx context.Context, id string) (*Saga, error) {
	data, err := s.client.Get(ctx, "saga:"+id).Bytes()
	if err != nil {
		return nil, err
	}
	var saga Saga
	if err := json.Unmarshal(data, &saga); err != nil {
		return nil, err
	}
	return &saga, nil
}
