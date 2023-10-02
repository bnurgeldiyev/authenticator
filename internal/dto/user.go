package dto

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"authenticator/config"
	"authenticator/internal/model"
)

type Create struct {
	Username string
	Password string
}

type AuthRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
}

type Validate struct {
	AccessToken string
}

type UpdateUsername struct {
	OldUsername string
	NewUsername string
}

type ChangeState struct {
	Username string
	State    model.State
}

type AuthTokenClaim struct {
	ID uuid.UUID
	jwt.StandardClaims
}

func GenerateAccessToken(id uuid.UUID) (accessToken string, err error) {

	ctx := context.Background()
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.dto").
		Str("method", "GenerateAccessToken").Logger()

	expiresAt := time.Now().Add(time.Minute * time.Duration(config.Conf.Jwt.AccessTokenExpiry)).Unix()

	claims := &AuthTokenClaim{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err = token.SignedString([]byte(config.Conf.Jwt.Secret))
	if err != nil {
		eMsg := "An error occurred on token.SignetString"
		zLog.Err(err).Msg(eMsg)
		return
	}

	return
}

func VerifyAccessToken(token string) (id uuid.UUID, err error) {

	ctx := context.Background()
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.dto").
		Str("method", "VerifyAccessToken").Logger()

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Conf.Jwt.Secret), nil
	})

	if err != nil {
		eMsg := "An error occurred on jwt.parse"
		zLog.Err(err).Msg(eMsg)
		return uuid.Nil, err
	}

	idStr := fmt.Sprintf("%v", claims["ID"])

	id = uuid.MustParse(idStr)

	return
}

func GenerateRefreshToken() (refreshToken uuid.UUID, err error) {
	ctx := context.Background()
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.dto").
		Str("method", "GenerateRefreshToken").Logger()

	refreshToken, err = uuid.NewUUID()
	if err != nil {
		eMsg := "An error occurred on uuid.NewUUID"
		zLog.Err(err).Msg(eMsg)
		return
	}

	return
}