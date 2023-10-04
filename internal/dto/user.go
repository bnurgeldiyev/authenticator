package dto

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

func NewGrpcError(err error) error {

	switch {
	case errors.Is(err, model.ErrUnauthorized):
		return status.Errorf(codes.Unauthenticated, "Unauthorized")
	case errors.Is(err, model.ErrNotFound):
		return status.Errorf(codes.NotFound, "Not found")
	case errors.Is(err, model.ErrConflict):
		return status.Errorf(codes.AlreadyExists, "Conflict")
	}

	return nil
}
