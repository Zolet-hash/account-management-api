package service

import (
	"context"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/domain"
	"github.com/Zolet-hash/account-management-api/internal/model"
	"github.com/google/uuid"
)

// and have the interface methods return those for specific error scenarios.

// --- Users Service Interface (Maps to 'Users' tag) ---

// UserService defines the methods for managing users and retrieving user-specific data.
type UserService interface {
	// CreateUser handles POST /users
	CreateUser(
		ctx context.Context,
		req model.CreateUserRequest,
	) (*model.CreateUserResponse, error) // returns user ID

	// GetUserBalances handles GET /users/{userId}/balances
	GetUserBalances(
		ctx context.Context,
		userID string,
	) (*model.GetUserBalancesResponse, error)

	// GetUserbalance handles GET /users/{userId}
}

// handles database operations
type UserStore interface {
	CreateUser(
		ctx context.Context,
		user model.CreateUserRequest,
	) (model.CreateUserResponse, error)

	GetUserBalances(
		ctx context.Context,
		userID uuid.UUID,
	) (model.GetUserBalancesResponse, error)
}

// Define service struct
type userService struct {
	store UserStore
	now   func() time.Time
}

// create constructor
func NewUserService(store UserStore) UserService {
	return &userService{
		store: store,
		now:   time.Now,
	}
}

// create user
func (s *userService) CreateUser(
	ctx context.Context,
	req model.CreateUserRequest,
) (*model.CreateUserResponse, error) {
	if req.Username == "" || req.DocumentNumber == "" || req.DateOfBirth.IsZero() ||
		req.DateOfBirth.After(s.now()) {
		return nil, domain.ErrInvalidRequest
	}

	user := model.CreateUserRequest{
		Username:       req.Username,
		DateOfBirth:    req.DateOfBirth,
		DocumentNumber: req.DocumentNumber,
	}

	createdUser, err := s.store.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &model.CreateUserResponse{
		UserID: createdUser.UserID,
	}, nil

}

// Get user balances
func (s *userService) GetUserBalances(
	ctx context.Context,
	userID string,
) (*model.GetUserBalancesResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	balances, err := s.store.GetUserBalances(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.GetUserBalancesResponse{
		TotalBalance:       balances.TotalBalance,
		IndividualBalances: balances.IndividualBalances,
	}, nil

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
