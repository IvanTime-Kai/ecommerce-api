package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyCache struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewIdempotencyCache(redis *redis.Client, ttl time.Duration) *IdempotencyCache {
	return &IdempotencyCache{
		redis: redis,
		ttl:   ttl,
	}
}

func (c *IdempotencyCache) IsProcessed(ctx context.Context, topic, eventID string) (bool, error) {
	key := fmt.Sprintf("idempotency:%s:%s", topic, eventID)
	exists, err := c.redis.Exists(ctx, key).Result()
	return exists > 0, err
}

func (c *IdempotencyCache) MarkProcessed(ctx context.Context, topic, eventID string) error {
	key := fmt.Sprintf("idempotency:%s:%s", topic, eventID)
	return c.redis.Set(ctx, key, 1, c.ttl).Err()
}
