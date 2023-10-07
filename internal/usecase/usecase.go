package usecase

import (
	"github.com/go-redis/redis/v8"

	"authenticator/internal/usecase/repo"
	"authenticator/internal/usecase/web"
	"authenticator/pkg/postgres"
)

type UseCases struct {
	UserUseCase *UserUseCase
}

func LoadUseCases(pg *postgres.Postgres, cache *redis.Client) *UseCases {
	txRepo := repo.NewTx(pg)
	userRepo := repo.NewUser(pg)
	w := web.NewWebAPI(cache)

	return &UseCases{
		UserUseCase: NewUserUseCase(userRepo, txRepo, w),
	}
}
