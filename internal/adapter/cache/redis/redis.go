package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const statusKeyPrefix = "msg_status:"

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

type StatusCache struct {
	client redisClient
}

func NewStatusCache(addr string) *StatusCache {
	return &StatusCache{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (c *StatusCache) GetStatus(ctx context.Context, id string) (string, error) {
	res, err := c.client.Get(ctx, statusKeyPrefix+id).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return res, nil
}

func (c *StatusCache) SetStatus(ctx context.Context, id, status string, ttl time.Duration) error {
	return c.client.Set(ctx, statusKeyPrefix+id, status, ttl).Err()
}

