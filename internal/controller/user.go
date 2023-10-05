package controller

import (
	"context"

	"github.com/rs/zerolog"

	"authenticator/internal/dto"
	"authenticator/internal/model"
	"authenticator/internal/usecase"
)

type UserRouter struct {
	u usecase.User
	AuthServiceServer
}

func NewUserRouter(u usecase.User) *UserRouter {
	return &UserRouter{
		u: u,
	}
}

func (r *UserRouter) Auth(ctx context.Context, in *AuthRequest) (*AuthResponse, error) {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.controller.User").
		Str("method", "Auth").Logger()

	authRequest := &dto.AuthRequest{
		Username: in.Username,
		Password: in.Password,
	}

	data, err := r.u.Auth(ctx, authRequest)
	if err != nil {
		zLog.Err(err).Msg("Error - r.u.Auth()")
		return nil, dto.NewGrpcError(err)
	}

	res := &AuthResponse{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
	}

	return res, nil
}

func (r *UserRouter) Create(ctx context.Context, in *CreateRequest) (*CreateResponse, error) {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.controller.User").
		Str("method", "Create").Logger()

	createRequest := &dto.Create{
		Username: in.Username,
		Password: in.Password,
	}

	err := r.u.Create(ctx, createRequest)
	if err != nil {
		zLog.Err(err).Msg("Error - Controller - User - create")
		return nil, dto.NewGrpcError(err)
	}

	return &CreateResponse{}, nil
}

func (r *UserRouter) Delete(ctx context.Context, in *DeleteRequest) (*DeleteResponse, error) {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.controller.User").
		Str("method", "Delete").Logger()

	changeStateRequest := &dto.ChangeState{
		Username: in.Username,
		State:    model.Deleted,
	}

	err := r.u.ChangeState(ctx, changeStateRequest)
	if err != nil {
		zLog.Err(err).Msg("Error - Controller - User - Delete")
		return nil, dto.NewGrpcError(err)
	}

	return &DeleteResponse{}, nil
}

func (r *UserRouter) ValidateToken(ctx context.Context, in *ValidateTokenRequest) (*ValidateTokenResponse, error) {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.controller.User").
		Str("method", "ValidateToken").Logger()

	err := r.u.Validate(ctx, in.AccessToken)
	if err != nil {
		zLog.Err(err).Msg("Error - Controller - User - ValidateToken")
		return nil, dto.NewGrpcError(err)
	}

	return &ValidateTokenResponse{}, nil
}

func (r *UserRouter) UpdateToken(ctx context.Context, in *UpdateTokenRequest) (*UpdateTokenResponse, error) {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.controller.User").
		Str("method", "UpdateToken").Logger()

	tokenRequest := &dto.UpdateToken{
		AccessToken:  in.AccessToken,
		RefreshToken: in.RefreshToken,
	}

	data, err := r.u.UpdateToken(ctx, tokenRequest)
	if err != nil {
		zLog.Err(err).Msg("Error - Controller - User - ValidateToken")
		return nil, dto.NewGrpcError(err)
	}

	res := &UpdateTokenResponse{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
	}

	return res, nil
}
