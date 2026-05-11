package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		key := fmt.Sprintf("rate_limit:%s", ip)

		now := time.Now().UnixMilli()

		windowStart := now - rl.window.Milliseconds()

		// Xóa timestamp cũ
		rl.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// Đếm request còn lại
		count, _ := rl.redis.ZCard(ctx, key).Result()

		// Nếu quá limit → trả 429
		if count >= int64(rl.limit) {
			http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"too many requests"}}`, http.StatusTooManyRequests)
			return
		}

		// Thêm timestamp mới
		rl.redis.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

		// Set TTL
		rl.redis.Expire(ctx, key, rl.window)

		next.ServeHTTP(w, r)
	})
}
