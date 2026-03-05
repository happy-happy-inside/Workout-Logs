package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	defaultTTL     time.Duration
	defaultTimeout time.Duration
	rdb            *redis.Client
}

func (r *RedisClient) checkTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), r.defaultTimeout)
	}

	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, r.defaultTimeout)
}

func NewRedisClient(defaultTimeout time.Duration, defaultTTL time.Duration, redisAddr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	redisClient := &RedisClient{
		defaultTimeout: defaultTimeout,
		rdb:            client,
		defaultTTL:     defaultTTL,
	}

	if err := redisClient.Ping(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

func (r *RedisClient) Ping() error {

	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)

		err := r.rdb.Ping(ctx).Err()
		cancel()

		if nil == err {
			return nil
		}

		log.Print(err)
	}

	return fmt.Errorf("Redis not reachebl")
}

func (r *RedisClient) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.defaultTTL
	}

	ctx, cancel := r.checkTimeout(ctx)
	defer cancel()

	if err := r.rdb.Set(ctx, key, val, ttl).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := r.checkTimeout(ctx)
	defer cancel()

	if res := r.rdb.Get(ctx, key).Val(); res != "" {
		return res, nil
	}

	return "", fmt.Errorf("Key not set")
}
