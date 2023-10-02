package model

import (
	"errors"
	"fmt"
)

type State string

var (
	ErrTypeNotMatched      = errors.New("type not matched")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("not found")
	ErrForbidden           = errors.New("forbidden")
	ErrAlreadyExists       = errors.New("already exists")
	ErrBadRequest          = errors.New("bad request")
	ErrConflict            = errors.New("conflict")
	ErrInternalServerError = errors.New("internal server error")
	ErrNoRowsAffected      = errors.New("no rows affected")
)

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"
	Deleted  State = "deleted"
)

func ParseState(s string) (r State, err error) {
	rt := State(s)
	switch rt {
	case Enabled,
		Disabled,
		Deleted:
		r = rt
		return
	default:
		return "", ErrTypeNotMatched
	}
}

type ErrUseCase struct {
	Err     error
	Message string
}

func (err ErrUseCase) Error() string {
	return fmt.Sprintf("%s %s", err.Message, err.Err)
}
