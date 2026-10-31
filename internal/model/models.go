package model

import "time"

// --- User Models ---

// CreateUserRequest is the request body for POST /users
type CreateUserRequest struct {
	Username       string    `json:"username"`
	DateOfBirth    time.Time `json:"dateOfBirth"`
	DocumentNumber string    `json:"documentNumber"`
}

// CreateUserResponse is the 201 response body for POST /users
type CreateUserResponse struct {
	UserID string `json:"userId"` // UUID
}

// UserBalance is an item in the IndividualBalances list
type UserBalance struct {
	ExchangeID string  `json:"exchangeId"`
	Balance    float64 `json:"balance"`
}

// GetUserBalancesResponse is the 200 response body for GET /users/{userId}/balances
type GetUserBalancesResponse struct {
	TotalBalance       float64       `json:"totalBalance"`
	IndividualBalances []UserBalance `json:"individualBalances"`
}

// --- Account Models ---

// CreateAccountRequest is the request body for POST /accounts
type CreateAccountRequest struct {
	UserID     string `json:"userId"` // UUID
	ExchangeID string `json:"exchangeId"`
}

// CreateAccountResponse is the 201 response body for POST /accounts
type CreateAccountResponse struct {
	AccountID string `json:"accountId"` // UUID
}

// --- Transaction Models ---

// DepositWithdrawalRequest is the request body for deposit/withdrawal endpoints
type DepositWithdrawalRequest struct {
	Amount float64 `json:"amount"` // min: 0.01
}

// TransactionResponse is the 200 response body for deposit/withdrawal endpoints
type TransactionResponse struct {
	AccountID  string  `json:"accountId"` // UUID
	NewBalance float64 `json:"newBalance"`
}

// ExchangeVolume is an item in the exchangeVolumes list
type ExchangeVolume struct {
	ExchangeID  string  `json:"exchangeId"`
	TotalVolume float64 `json:"totalVolume"`
}

// DailyVolume is an item in the dailyVolumes list
type DailyVolume struct {
	Date            time.Time        `json:"date"`
	ExchangeVolumes []ExchangeVolume `json:"exchangeVolumes"`
}

// GetTransactionVolumeResponse is the 200 response body for GET /transactions/volume
type GetTransactionVolumeResponse struct {
	DailyVolumes []DailyVolume `json:"dailyVolumes"`
}
