package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Delete(ctx context.Context, key string) error
}

type redisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, db int, password string, ttl time.Duration) Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisCache{
		client: rdb,
		ttl:    ttl,
	}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCache) Set(ctx context.Context, key, value string) error {
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
