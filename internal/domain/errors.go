package domain

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrTransferLimitExceeded = errors.New("transfer limit exceeded")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrIdemptotencyConflict  = errors.New("idemptotency key conflict")
)
