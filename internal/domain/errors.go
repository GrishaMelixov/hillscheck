package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrDuplicateExternal = errors.New("transaction with this external_id already exists")
	ErrInvalidAmount     = errors.New("amount must be non-zero integer cents")
	ErrPoolFull          = errors.New("worker pool queue is full")
	ErrUnauthorized      = errors.New("unauthorized")
)
