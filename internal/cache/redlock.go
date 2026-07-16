package cache

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	goredislib "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredis "github.com/redis/go-redis/v9"
)

type RedLockClient struct {
	rs *redsync.Redsync
}

func NewRedLockClient(urls []string) *RedLockClient {
	var pools []redsyncredis.Pool

	for _, url := range urls {
		client := goredis.NewClient(&goredis.Options{Addr: url})
		pools = append(pools, goredislib.NewPool(client))
	}

	return &RedLockClient{rs: redsync.New(pools...)}
}

func (r *RedLockClient) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*redsync.Mutex, bool) {

	mutex := r.rs.NewMutex(key, redsync.WithExpiry(ttl), redsync.WithTries(1))

	if err := mutex.LockContext(ctx); err != nil {
		return nil, false
	}

	return mutex, true
}

func (r *RedLockClient) ReleaseLock(ctx context.Context, mutex *redsync.Mutex) {
	mutex.UnlockContext(ctx)
}
