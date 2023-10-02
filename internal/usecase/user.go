package usecase

import (
	"context"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"authenticator/internal/dto"
	"authenticator/internal/model"
	"authenticator/pkg/util"
)

// UserUseCase -.
type UserUseCase struct {
	repo   UserRepo
	txRepo TxRepo
}

// NewUserUseCase -.
func NewUserUseCase(r UserRepo, tx TxRepo) *UserUseCase {
	return &UserUseCase{
		repo:   r,
		txRepo: tx,
	}
}

func (uc *UserUseCase) Auth(ctx context.Context, in *dto.AuthRequest) (*dto.AuthResponse, error) {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.UserUseCase").
		Str("method", "Login").Logger()

	user, err := uc.repo.GetByUsername(ctx, in.Username)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing uc.repo.GetByUsername")
		return nil, err
	}

	if user == nil {
		return nil, &model.ErrUseCase{
			Err:     model.ErrUnauthorized,
			Message: "user not found",
		}
	}
	defer func() {
		if err != nil {
			user = nil
		}
	}()

	if err = util.VerifyPasswordFromHash(in.Password, user.Password); err != nil {
		zLog.Err(err).Msg("error verifying password")

		if err == bcrypt.ErrMismatchedHashAndPassword {
			err = &model.ErrUseCase{
				Err:     model.ErrUnauthorized,
				Message: "password hash mismatch",
			}
		}

		return nil, err
	}

	accessToken, err := dto.GenerateAccessToken(user.Id)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing dto.GenerateAccessToken()")
		return nil, err
	}

	refreshToken, err := dto.GenerateRefreshToken()
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing dto.GenerateRefreshToken()")
		return nil, err
	}

	item := &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String(),
	}

	return item, nil
}

func (uc *UserUseCase) Create(ctx context.Context, in *dto.Create) error {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.UserUseCase").
		Str("method", "Login").Logger()

	user, err := uc.repo.GetByUsername(ctx, in.Username)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing uc.repo.GetByUsername")
		return err
	}

	if user != nil {
		return &model.ErrUseCase{
			Err:     model.ErrConflict,
			Message: "conflict",
		}
	}

	pwdHash, err := util.HashPassword(in.Password)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error util.HashPassword")
		return err
	}

	var txId int
	txId, err = uc.txRepo.NewTxId(ctx)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing r.txRepo.NewTxId")
		return err
	}
	defer func() {
		err = uc.txRepo.TxEnd(ctx, txId, err)
		if err != nil {
			zLog.Err(err).Msg("UserUseCase - error processing r.txRepo.TxEnd")
			return
		}
	}()

	now := util.NowUTC()
	userModel := &model.User{
		Username: in.Username,
		Password: pwdHash,
		State:    model.Enabled,
		CreateTs: now,
		UpdateTs: now,
	}

	err = uc.repo.Create(ctx, userModel, txId)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing d.repo.Create")
		return err
	}

	return nil
}