package service

import (
	"context"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/model"
)

// and have the interface methods return those for specific error scenarios.

// --- Users Service Interface (Maps to 'Users' tag) ---

// UserService defines the methods for managing users and retrieving user-specific data.
type UserService interface {
	// CreateUser handles POST /users
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.CreateUserResponse, error)

	// GetUserBalances handles GET /users/{userId}/balances
	GetUserBalances(ctx context.Context, userID string) (*model.GetUserBalancesResponse, error)
}

// --- Accounts Service Interface (Maps to 'Accounts' tag) ---

// AccountService defines the methods for creating accounts.
type AccountService interface {
	// CreateAccount handles POST /accounts
	CreateAccount(ctx context.Context, req model.CreateAccountRequest) (*model.CreateAccountResponse, error)
}

// --- Transactions Service Interface (Maps to 'Transactions' tag) ---

// TransactionService defines the methods for managing financial movements and volume reporting.
type TransactionService interface {
	// Deposit handles POST /accounts/{accountId}/deposit
	Deposit(ctx context.Context, accountID string, req model.DepositWithdrawalRequest) (*model.TransactionResponse, error)

	// Withdraw handles POST /accounts/{accountId}/withdrawal
	Withdraw(ctx context.Context, accountID string, req model.DepositWithdrawalRequest) (*model.TransactionResponse, error)

	// GetDailyTransactionVolume handles GET /transactions/volume
	GetDailyTransactionVolume(ctx context.Context, startDate time.Time, endDate time.Time) (*model.GetTransactionVolumeResponse, error)
}
