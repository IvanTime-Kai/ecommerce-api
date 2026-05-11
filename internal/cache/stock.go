package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var deductStockScript = redis.NewScript(`
	local stock = redis.call('GET', KEYS[1])
	if stock == false or tonumber(stock) < tonumber(ARGV[1]) then
		return 0
	end
	redis.call('DECRBY', KEYS[1], ARGV[1])
	return 1
`)

type StockCache struct {
	redis *redis.Client
}

func NewStockCache(redis *redis.Client) *StockCache {
	return &StockCache{
		redis: redis,
	}
}

func (s *StockCache) DeductStock(ctx context.Context, productID string, quantity int) (bool, error) {

	key := fmt.Sprintf("stock:%s", productID)

	result, err := deductStockScript.Run(ctx, s.redis, []string{key}, quantity).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (s *StockCache) SetStock(ctx context.Context, productID string, stock int) error {
	key := fmt.Sprintf("stock:%s", productID)
	return s.redis.Set(ctx, key, stock, 0).Err()
}
