package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sunxi/92-Account-Center/session-service/internal/model"
)

var ErrSessionNotFound = errors.New("session not found")

const (
	SessionKeyPrefix      = "session:"
	UserSessionsKeyPrefix = "user_sessions:"
	SessionTTL            = 20 * time.Minute
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	CreateWithEviction(ctx context.Context, session *model.Session, maxSessions int64) error
	GetByID(ctx context.Context, sessionID string) (*model.Session, error)
	GetUserSessions(ctx context.Context, userID int64) ([]*model.Session, error)
	UpdateLastAccessed(ctx context.Context, sessionID string, lastAccessedAt time.Time, expiresAt time.Time) error
	Delete(ctx context.Context, sessionID string, userID int64) error
	DeleteAllUserSessions(ctx context.Context, userID int64) error
	CountUserSessions(ctx context.Context, userID int64) (int64, error)
	GetSessionTTL(ctx context.Context, sessionID string) (time.Duration, error)
}

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return &sessionRepository{rdb: rdb}
}

func (r *sessionRepository) sessionKey(sessionID string) string {
	return SessionKeyPrefix + sessionID
}

func (r *sessionRepository) userSessionsKey(userID int64) string {
	return fmt.Sprintf("%s%d", UserSessionsKeyPrefix, userID)
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	pipe := r.rdb.Pipeline()

	pipe.Set(ctx, r.sessionKey(session.SessionID), data, SessionTTL)

	pipe.SAdd(ctx, r.userSessionsKey(session.UserID), session.SessionID)

	pipe.Expire(ctx, r.userSessionsKey(session.UserID), SessionTTL*10)

	_, err = pipe.Exec(ctx)
	return err
}

var createWithEvictionScript = redis.NewScript(`
local userSessionsKey = KEYS[1]
local sessionKeyPrefix = "session:"
local maxSessions = tonumber(ARGV[1])
local sessionData = ARGV[2]
local sessionTTL = tonumber(ARGV[3])
local sessionID = ARGV[4]

local sessionIDs = redis.call("SMEMBERS", userSessionsKey)
local count = 0
local oldestSessionID = nil
local oldestTime = nil

for i, sid in ipairs(sessionIDs) do
    local exists = redis.call("EXISTS", sessionKeyPrefix .. sid)
    if exists == 1 then
        count = count + 1
        local data = redis.call("GET", sessionKeyPrefix .. sid)
        if data then
            local session = cjson.decode(data)
            local createdAt = session.CreatedAt
            if oldestTime == nil or createdAt < oldestTime then
                oldestTime = createdAt
                oldestSessionID = sid
            end
        end
    else
        redis.call("SREM", userSessionsKey, sid)
    end
end

if count >= maxSessions and oldestSessionID then
    redis.call("DEL", sessionKeyPrefix .. oldestSessionID)
    redis.call("SREM", userSessionsKey, oldestSessionID)
end

redis.call("SET", sessionKeyPrefix .. sessionID, sessionData, sessionTTL)
redis.call("SADD", userSessionsKey, sessionID)
redis.call("EXPIRE", userSessionsKey, sessionTTL * 10)

return 1
`)

func (r *sessionRepository) CreateWithEviction(ctx context.Context, session *model.Session, maxSessions int64) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttlSeconds := int64(SessionTTL.Seconds())
	if ttlSeconds <= 0 {
		ttlSeconds = 1200
	}

	_, err = createWithEvictionScript.Run(ctx, r.rdb, []string{
		r.userSessionsKey(session.UserID),
	}, maxSessions, string(data), ttlSeconds*1000000000, session.SessionID).Result()

	return err
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID string) (*model.Session, error) {
	data, err := r.rdb.Get(ctx, r.sessionKey(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var session model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]*model.Session, error) {
	sessionIDs, err := r.rdb.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	var sessions []*model.Session
	for _, sessionID := range sessionIDs {
		session, err := r.GetByID(ctx, sessionID)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				continue
			}
			return nil, err
		}
		if session.IsActive {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

func (r *sessionRepository) UpdateLastAccessed(ctx context.Context, sessionID string, lastAccessedAt time.Time, expiresAt time.Time) error {
	session, err := r.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastAccessedAt = lastAccessedAt
	session.ExpiresAt = expiresAt

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl, err := r.rdb.TTL(ctx, r.sessionKey(sessionID)).Result()
	if err != nil {
		return err
	}

	if ttl > 0 {
		return r.rdb.Set(ctx, r.sessionKey(sessionID), data, ttl).Err()
	}

	return r.rdb.Set(ctx, r.sessionKey(sessionID), data, SessionTTL).Err()
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string, userID int64) error {
	pipe := r.rdb.Pipeline()

	pipe.Del(ctx, r.sessionKey(sessionID))

	pipe.SRem(ctx, r.userSessionsKey(userID), sessionID)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *sessionRepository) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	sessionIDs, err := r.rdb.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := r.rdb.Pipeline()

	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, r.sessionKey(sessionID))
	}

	pipe.Del(ctx, r.userSessionsKey(userID))

	_, err = pipe.Exec(ctx)
	return err
}

func (r *sessionRepository) CountUserSessions(ctx context.Context, userID int64) (int64, error) {
	sessionIDs, err := r.rdb.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return 0, err
	}

	var count int64
	for _, sessionID := range sessionIDs {
		exists, err := r.rdb.Exists(ctx, r.sessionKey(sessionID)).Result()
		if err != nil {
			return 0, err
		}
		if exists > 0 {
			count++
		} else {
			r.rdb.SRem(ctx, r.userSessionsKey(userID), sessionID)
		}
	}

	return count, nil
}

func (r *sessionRepository) GetSessionTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	return r.rdb.TTL(ctx, r.sessionKey(sessionID)).Result()
}
