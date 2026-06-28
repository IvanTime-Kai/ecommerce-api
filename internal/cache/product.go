package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type ProductCache struct {
	redis *redis.Client
}

func NewProductCache(redis *redis.Client) *ProductCache {
	return &ProductCache{
		redis: redis,
	}
}

func (p *ProductCache) DeleteByPattern(ctx context.Context, pattern string) error {

	var cursor uint64

	for {
		keys, nextCursor, err := p.redis.Scan(ctx, cursor, pattern, 100).Result()

		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := p.redis.Unlink(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}

	}

	return nil
}