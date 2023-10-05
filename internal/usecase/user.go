package usecase

import (
	"context"
	"errors"
	"fmt"

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
	webAPI WebAPI
}

// NewUserUseCase -.
func NewUserUseCase(r UserRepo, tx TxRepo, w WebAPI) *UserUseCase {
	return &UserUseCase{
		repo:   r,
		txRepo: tx,
		webAPI: w,
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
		eMsg := fmt.Sprintf("User with username = <%s> not found", in.Username)
		zLog.Err(fmt.Errorf("user not found")).Msg(eMsg)
		return nil, model.ErrUnauthorized
	}
	defer func() {
		if err != nil {
			user = nil
		}
	}()

	if err = util.VerifyPasswordFromHash(in.Password, user.Password); err != nil {
		zLog.Err(err).Msg("error verifying password")

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			err = model.ErrUnauthorized
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

	err = uc.webAPI.AddRefreshToken(ctx, user.Id, refreshToken.String())
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error uc.webAPI.AddRefreshToken")
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
		Str("method", "Create").Logger()

	user, err := uc.repo.GetByUsername(ctx, in.Username)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing uc.repo.GetByUsername")
		return err
	}

	if user != nil {
		return model.ErrConflict
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

func (uc *UserUseCase) ChangeState(ctx context.Context, in *dto.ChangeState) error {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.UserUseCase").
		Str("method", "ChangeState").Logger()

	user, err := uc.repo.GetByUsername(ctx, in.Username)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing uc.repo.GetByUsername")
		return err
	}

	if user == nil {
		return model.ErrNotFound
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

	userModel := &model.User{
		State:    in.State,
		UpdateTs: util.NowUTC(),
		Version:  util.VersionInc(user.Version),
	}

	err = uc.repo.ChangeState(ctx, user, userModel, txId)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing d.repo.ChangeState")
		return err
	}

	return nil
}

func (uc *UserUseCase) Validate(ctx context.Context, token string) error {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.UserUseCase").
		Str("method", "Validate").Logger()

	userId, err := dto.VerifyAccessToken(token)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error dto.VerifyAccessToken")
		return model.ErrUnauthorized
	}

	_, err1 := uc.repo.GetById(ctx, userId)
	if err1 != nil {
		zLog.Err(err).Msg("UserUseCase - error uc.repo.GetById")
		return err1
	}

	return nil
}

func (uc *UserUseCase) UpdateToken(ctx context.Context, in *dto.UpdateToken) (*dto.UpdateToken, error) {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.UserUseCase").
		Str("method", "UpdateToken").Logger()

	userId, err := dto.VerifyAccessToken(in.AccessToken)
	if err != nil && err.Error() != "Token is expired" {
		if err.Error() != "Token is expired" {
			zLog.Err(err).Msg("UserUseCase - error dto.VerifyAccessToken")
			return nil, model.ErrUnauthorized
		}
	}

	if err == nil {
		zLog.Info().Msg("UserUseCase - info - accessToken is not expired")
		return nil, model.ErrForbidden
	}

	userById, err1 := uc.repo.GetById(ctx, userId)
	if err1 != nil {
		zLog.Err(err).Msg("UserUseCase - error uc.repo.GetById")
		return nil, err
	}

	if userById == nil {
		eMsg := fmt.Sprintf("User with id = <%s> not found", userId)
		zLog.Err(nil).Msg(eMsg)
		return nil, err
	}

	accessToken, err := dto.GenerateAccessToken(userId)
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing dto.GenerateAccessToken()")
		return nil, err
	}

	refreshToken, err := dto.GenerateRefreshToken()
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error processing dto.GenerateRefreshToken()")
		return nil, err
	}

	err = uc.webAPI.AddRefreshToken(ctx, userId, refreshToken.String())
	if err != nil {
		zLog.Err(err).Msg("UserUseCase - error uc.webAPI.AddRefreshToken")
		return nil, err
	}

	item := &dto.UpdateToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String(),
	}

	fmt.Println(item)

	return item, nil
}
