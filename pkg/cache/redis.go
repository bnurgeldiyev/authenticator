package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
)

func NewRedisService(ctx context.Context, redisConn string, redisDb, maxIdle, maxActive int) (r *redis.Client, err error) {

	client := redis.NewClient(&redis.Options{
		Addr:         redisConn,
		DB:           redisDb,
		PoolSize:     maxActive,
		MinIdleConns: maxIdle,
	})

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return client, nil
}
