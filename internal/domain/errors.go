package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrDuplicateExternal = errors.New("transaction with this external_id already exists")
	ErrInvalidAmount     = errors.New("amount must be non-zero integer cents")
	ErrPoolFull          = errors.New("worker pool queue is full")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrEmailTaken        = errors.New("email already registered")
	ErrWrongPassword     = errors.New("invalid email or password")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
)
