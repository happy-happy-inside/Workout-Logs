package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	defaultTimeout time.Duration
	rdb            *redis.Client
	mu             sync.RWMutex
}

func NewRedisClient(defaultTimeout time.Duration, redisAddr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	redisClient := &RedisClient{
			defaultTimeout: defaultTimeout,
		rdb:            client,
	}

	if err := redisClient.Ping(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

func (r *RedisClient) Ping() error {
	
	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
		defer cancel()

		err := r.rdb.Ping(ctx).Err()
		if nil == err {
			return nil
		}
		log.Print(err)
	}
	
	return fmt.Errorf("Redis not reachebl")
}

func (r *RedisClient) Set(ctx context.Context, key string, val string) error {

}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {

}
