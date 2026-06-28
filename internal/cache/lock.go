package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LockCache struct {
	redis *redis.Client
}

var releaseLockScript = redis.NewScript(`
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    end
    return 0
`)

func NewLockCache(redis *redis.Client) *LockCache {
	return &LockCache{
		redis: redis,
	}
}

func (l *LockCache) AcquireLock(ctx context.Context, key string, instanceID string, ttl time.Duration) bool {
	result, err := l.redis.SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return false
	}
	return result
}

func (l *LockCache) ReleaseLock(ctx context.Context, key string, instanceID string) error {
	return releaseLockScript.Run(ctx, l.redis, []string{key}, instanceID).Err()
}
