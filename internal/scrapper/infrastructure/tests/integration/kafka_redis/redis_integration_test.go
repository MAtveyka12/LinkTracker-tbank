package kafkaredis_test

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/cache"
)

func (s *TestSuite) TestRedisCache_SetGetDelete() {
	ctx := context.Background()
	redisCache := cache.NewRedisCache(s.redisContainer.Addr, 0, "", 1*time.Minute)

	key := "integration:test:key"
	val := "test-value"

	err := redisCache.Set(ctx, key, val)
	s.Require().NoError(err)

	got, err := redisCache.Get(ctx, key)
	s.Require().NoError(err)
	s.Equal(val, got)

	err = redisCache.Delete(ctx, key)
	s.Require().NoError(err)

	_, err = redisCache.Get(ctx, key)
	s.Require().Error(err)
}

func (s *TestSuite) TestCacheInvalidation() {
	ctx := context.Background()
	redisCache := cache.NewRedisCache(s.redisContainer.Addr, 0, "", 1*time.Minute)

	key := "test:cache:key"
	val := "value"

	s.Require().NoError(redisCache.Set(ctx, key, val))

	got, err := redisCache.Get(ctx, key)
	s.Require().NoError(err)
	s.Equal(val, got)

	s.Require().NoError(redisCache.Delete(ctx, key))

	_, err = redisCache.Get(ctx, key)
	s.Require().Error(err)
}

func (s *TestSuite) TestRedisConnectivity() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr: s.redisContainer.Addr,
	})
	defer client.Close()

	status := client.Ping(ctx)
	s.Require().NoError(status.Err(), "Failed to connect to Redis")
}

func (s *TestSuite) TestRedisExpiration() {
	ctx := context.Background()
	redisCache := cache.NewRedisCache(s.redisContainer.Addr, 0, "", 1*time.Second)

	key := "test:expiration:key"
	val := "expiring-value"

	err := redisCache.Set(ctx, key, val)
	s.Require().NoError(err)

	got, err := redisCache.Get(ctx, key)
	s.Require().NoError(err)
	s.Equal(val, got)

	time.Sleep(2 * time.Second)

	_, err = redisCache.Get(ctx, key)
	s.Require().Error(err)
}

func (s *TestSuite) TestRedisConcurrentAccess() {
	ctx := context.Background()
	redisCache := cache.NewRedisCache(s.redisContainer.Addr, 0, "", 1*time.Minute)

	key1 := "test:concurrent:key1"
	key2 := "test:concurrent:key2"
	val1 := "value1"
	val2 := "value2"

	err1 := make(chan error, 1)
	err2 := make(chan error, 1)

	go func() {
		err1 <- redisCache.Set(ctx, key1, val1)
	}()

	go func() {
		err2 <- redisCache.Set(ctx, key2, val2)
	}()

	s.Require().NoError(<-err1)
	s.Require().NoError(<-err2)

	got1, err := redisCache.Get(ctx, key1)
	s.Require().NoError(err)
	s.Equal(val1, got1)

	got2, err := redisCache.Get(ctx, key2)
	s.Require().NoError(err)
	s.Equal(val2, got2)

	s.Require().NoError(redisCache.Delete(ctx, key1))
	s.Require().NoError(redisCache.Delete(ctx, key2))
}
