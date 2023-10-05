package web

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"authenticator/config"
)

type WebAPI struct {
	cache *redis.Client
}

func NewWebAPI(cache *redis.Client) *WebAPI {
	return &WebAPI{
		cache: cache,
	}
}

const userRefreshToken = "user:refreshToken"

func (w *WebAPI) AddRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) (err error) {

	key := fmt.Sprintf("%v:%v", userRefreshToken, id)

	err = w.cache.Set(ctx, key, refreshToken, time.Duration(config.Conf.Jwt.RefreshTokenExpiry)*time.Minute).Err()

	return
}

func (w *WebAPI) GetRefreshTokenByID(ctx context.Context, id uuid.UUID) (refreshToken uuid.UUID, err error) {

	key := fmt.Sprintf("%v:%v", userRefreshToken, id)

	tokenStr, err := w.cache.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return
	}

	if err == redis.Nil {
		return uuid.Nil, nil
	}

	refreshToken = uuid.MustParse(tokenStr)

	return
}
