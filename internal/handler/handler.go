package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/domain"
	"github.com/Zolet-hash/account-management-api/internal/service"
	"github.com/gin-gonic/gin"
)

// initialize a new handler func
type Handler struct {
	s *service.Service
}

// New creates and returns new handlers
func New(s *service.Service) *Handler {
	return &Handler{s: s}
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// respond to err with detailed err message
func respondError(c *gin.Context, err error) {
	status, code, msg := http.StatusInternalServerError, "internal_error", "An unexpected error occurred"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, msg = 404, "not_found", "Resource not found"
	case errors.Is(err, domain.ErrConflict):
		status, code, msg = 409, "conflict", "Resource already exists"
	case errors.Is(err, domain.ErrIdempotencyConflict):
		status, code, msg = 409, "idempotency_conflict", "Idempotency key was used for a different operation"
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidRequest):
		status, code, msg = 400, "invalid_request", "Request is invalid"
	case errors.Is(err, domain.ErrInsufficientBalance):
		status, code, msg = 422, "insufficient_balance", "Insufficient balance"
	case errors.Is(err, domain.ErrTransferLimitExceeded):
		status, code, msg = 422, "transfer_limit_exceeded", "Transfer limit exceeded"
	case errors.Is(err, domain.ErrMinimumAgeNotMet):
		status, code, msg = 422, "minimum_age_not_met", "User does not meet exchange minimum age"
	case errors.Is(err, contextCanceled), errors.Is(err, context.DeadlineExceeded):
		status, code, msg = 408, "request_cancelled", "Request was cancelled"
	}
	var out errorResponse
	out.Error.Code = code
	out.Error.Message = msg
	c.JSON(status, out)
}

// context cancelled err
var contextCanceled = context.Canceled

// User creatin handler
func (h *Handler) CreateUser(c *gin.Context) {
	var b struct {
		Username       string `json:"username" binding:"required,max=100"`
		DateOfBirth    string `json:"dateOfBirth" binding:"required"`
		DocumentNumber string `json:"documentNumber" binding:"required,max=100"`
	}
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, domain.ErrInvalidRequest)
		return
	}
	d, e := time.Parse("2006-01-02", b.DateOfBirth)
	if e != nil {
		respondError(c, domain.ErrInvalidRequest)
		return
	}
	u, e := h.s.CreateUser(c.Request.Context(), b.Username, d, b.DocumentNumber)
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(201, u)
}
