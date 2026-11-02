package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

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

// User creatin Use handler
func (handler *Handler) CreateUser(ctx *gin.Context) {
	var userRequest struct {
		Username       string `json:"username" binding:"required,max=100"`
		DateOfBirth    string `json:"dateOfBirth" binding:"required"`
		DocumentNumber string `json:"documentNumber" binding:"required,max=100"`
	}
	if ctx.ShouldBindJSON(&userRequest) != nil {
		respondError(ctx, domain.ErrInvalidRequest)
		return
	}
	dob, err := time.Parse("2006-01-02", userRequest.DateOfBirth) // go uses 2006-01-02 instead of YYYY-MM-DD
	if err != nil {
		respondError(ctx, domain.ErrInvalidRequest)
		return
	}
	user, err := handler.s.CreateUser(ctx.Request.Context(), userRequest.Username, dob, userRequest.DocumentNumber)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(201, user)
}

// create account
func (handler *Handler) CreateAccount(ctx *gin.Context) {

	var account struct {
		UserID     string `json:"userId" binding:"required,uuid"`
		ExchangeID string `json:"exchangeId" binding:"required,uuid"`
	}

	if ctx.ShouldBindJSON(&account) != nil {
		respondError(ctx, domain.ErrInvalidRequest)
		return
	}

	userId, _ := uuid.Parse(account.UserID)
	exchangeId, _ := uuid.Parse(account.ExchangeID)
	accountId, err := handler.s.CreateAccount(ctx.Request.Context(), userId, exchangeId)

	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(201, accountId)
}

// transfer
func (handler *Handler) transfer(typ string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("accountId"))
		if err != nil {
			respondError(ctx, domain.ErrInvalidRequest)
			return
		}
		var amount struct {
			Amount float64 `json:"amount" binding:"required"`
		}
		if ctx.ShouldBindJSON(&amount) != nil {
			respondError(ctx, domain.ErrInvalidRequest)
			return
		}
		key := ctx.GetHeader("Idempotency-Key")
		if key == "" {
			key = ctx.Query("idempotencyKey")
		}
		trans, err := handler.s.Transfer(ctx.Request.Context(), id, typ, amount.Amount, key)
		if err != nil {
			respondError(ctx, err)
			return
		}
		ctx.JSON(200, trans)
	}
}

// User balances
func (handler *Handler) Balances(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		respondError(ctx, domain.ErrInvalidRequest)
		return
	}
	balance, err := handler.s.Balances(ctx.Request.Context(), id)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(200, balance)
}

func (h *Handler) Volumes(ctx *gin.Context) {
	start, e1 := time.Parse("2006-01-02", ctx.Query("startDate"))
	end, e2 := time.Parse("2006-01-02", ctx.Query("endDate"))
	if e1 != nil || e2 != nil {
		respondError(ctx, domain.ErrInvalidRequest)
		return
	}
	volume, err := h.s.Volumes(ctx.Request.Context(), start, end)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(200, volume)
}

func (h *Handler) TransferDeposit() gin.HandlerFunc {
	return h.transfer("deposit")
}

func (h *Handler) TransferWithdrawal() gin.HandlerFunc {
	return h.transfer("withdrawal")
}
