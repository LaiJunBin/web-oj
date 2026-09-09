package usecase

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("expired token")
)
