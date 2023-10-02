package usecase

import (
	"authenticator/internal/usecase/repo"
	"authenticator/pkg/postgres"
)

type UseCases struct {
	UserUseCase *UserUseCase
}

func LoadUseCases(pg *postgres.Postgres) *UseCases {
	txRepo := repo.NewTx(pg)
	userRepo := repo.NewUser(pg)

	return &UseCases{
		UserUseCase: NewUserUseCase(userRepo, txRepo),
	}
}
