package usecase

import (
	"context"

	"github.com/google/uuid"

	"authenticator/internal/dto"
	"authenticator/internal/model"
)

type (
	User interface {
		Auth(ctx context.Context, in *dto.AuthRequest) (*dto.AuthResponse, error)
		Create(ctx context.Context, in *dto.Create) error
	}
)

type (
	TxRepo interface {
		NewTxId(ctx context.Context) (txId int, err error)
		TxEnd(ctx context.Context, txId int, err error) error
	}

	UserRepo interface {
		Create(ctx context.Context, in *model.User, txId int) error
		GetById(ctx context.Context, id uuid.UUID) (*model.User, error)
		GetByUsername(ctx context.Context, username string) (*model.User, error)
		GetPasswordById(ctx context.Context, id uuid.UUID) (*model.User, error)
		ChangeState(ctx context.Context, old, new *model.User, txId int) error
	}
)
