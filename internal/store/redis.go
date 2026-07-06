package store

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(addr string) *Redis {
	addr = os.Getenv("REDIS_ADDR")

	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Redis{Client: rdb}
}

func (r *Redis) SetJTI(ctx context.Context, key, userID string, exp time.Time) error {
	return r.Client.Set(ctx, key, userID, time.Until(exp)).Err()
}

func (r *Redis) DelJTI(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *Redis) GetUserByJTI(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}
