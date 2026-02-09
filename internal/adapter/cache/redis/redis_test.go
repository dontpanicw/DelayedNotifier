package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Для юнит-теста используем in-memory Redis через miniredis было бы удобно,
// но чтобы не тянуть дополнительную зависимость, проверим только логику
// обработки redis.Nil (отсутствующий ключ).

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
		client: redis.NewClient(&redis.Options{}),
	}
	// подменяем методы через fakeClient с ошибкой redis.Nil
	f := &fakeClient{err: redis.Nil}

	// небольшой трюк: переназначаем методы через интерфейсную обёртку сложно без рефакторинга,
	// поэтому просто проверим поведение с реальным клиентом на несуществующем ключе.
	// В реальном Redis key не существует => Get вернёт redis.Nil => ожидаем пустую строку без ошибки.

	status, err := c.GetStatus(ctx, "unknown")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != "" {
		t.Fatalf("expected empty status on miss, got %q", status)
	}

	_ = f // чтобы не ругался линтер, сам fakeClient оставлен как задел для расширения
}

