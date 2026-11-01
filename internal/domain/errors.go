package domain

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrTransferLimitExceeded = errors.New("transfer limit exceeded")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrIdempotencyConflict  = errors.New("idemptotency key conflict")
	ErrMinimumAgeNotMet      = errors.New("minimum age not met")
)
