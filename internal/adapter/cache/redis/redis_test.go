package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// fakeClient реализует минимальный интерфейс redisClient и позволяет
// тестировать поведение без реального Redis.
type fakeClient struct {
	value string
	err   error
}

func (f *fakeClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	if f.err != nil {
		cmd.SetErr(f.err)
		return cmd
	}
	cmd.SetVal(f.value)
	return cmd
}

func (f *fakeClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetVal("OK")
	return cmd
}

func TestStatusCache_GetStatus_MissReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	c := &StatusCache{
		client: &fakeClient{err: redis.Nil},
	}

	status, err := c.GetStatus(ctx, "unknown")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "" {
		t.Fatalf("expected empty status on miss, got %q", status)
	}
}

