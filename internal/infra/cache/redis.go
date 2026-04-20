package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = 7 * 24 * time.Hour
const tiposAtoKey = "tipos_ato"

type Cache struct {
	client *redis.Client
}

func New(addr string) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis indisponível em %s: %w", addr, err)
	}
	return &Cache{client: rdb}, nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key, value string) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) GetTiposAto(ctx context.Context) (string, error) {
	return c.client.Get(ctx, tiposAtoKey).Result()
}

func (c *Cache) SetTiposAto(ctx context.Context, value string) error {
	return c.client.Set(ctx, tiposAtoKey, value, 0).Err()
}

func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
