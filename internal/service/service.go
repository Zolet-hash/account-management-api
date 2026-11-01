package service

import (
	"context"
	"math"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/domain"
	"github.com/Zolet-hash/account-management-api/internal/model"
	"github.com/google/uuid"
)

type Store interface {
	CreateUser(context.Context, model.User) (model.User, error)
	User(context.Context, uuid.UUID) (model.User, error)
	Exchange(context.Context, uuid.UUID) (model.Exchange, error)
	CreateAccount(context.Context, model.Account) (model.Account, error)
	Account(context.Context, uuid.UUID) (model.Account, model.Exchange, error)
	Transfer(context.Context, uuid.UUID, string, int64, string) (model.Transaction, error)
	Balances(context.Context, uuid.UUID) (model.UserBalances, error)
	Volumes(context.Context, time.Time, time.Time) ([]model.DailyVolume, error)
}

// and have the interface methods return those for specific error scenarios.

type Service struct {
	store Store
	now   func() time.Time
}

func New(s Store) *Service {
	return &Service{
		store: s,
		now:   time.Now,
	}
}

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

// create user
func (s *Service) CreateUser(ctx context.Context, username string, dob time.Time, document string) (*model.CreateUserResponse, error) {
	if username == "" || document == "" || dob.IsZero() ||
		dob.After(s.now()) {
		return nil, domain.ErrInvalidRequest
	}

	return s.store.CreateUser(ctx, model.CreateUserRequest{Username: username, DateOfBirth: dob, DocumentNumber: document})

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
	CreateAccount(
		ctx context.Context,
		req model.CreateAccountRequest,
	) (*model.CreateAccountResponse, error) // returns acc id
}

type AccountStore interface {
	CreateAccount(
		ctx context.Context,
		account model.CreateAccountRequest,
	) (*model.CreateAccountResponse, error) // returns account ID
}

type accountService struct {
	store AccountStore
	now   func() time.Time
}

func NewAccountService(store AccountStore) AccountService {
	return &accountService{
		store: store,
		now:   time.Now,
	}
}

// create account
func (s *accountService) CreateAccount(
	ctx context.Context,
	req model.CreateAccountRequest,
) (*model.CreateAccountResponse, error) { // returns acc id

	//retrieve the user
	user, err := s.store.userID(ctx, userID)

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

// calculateAge calculates a person's age in complete years based on their DOB
// also handles pple born on feb 29
func calculateAge(dob, current time.Time) int {
	age := current.Year() - dob.Year()

	// if current date is nov 1st 2026 and DOB is June 21 2000 -> bithday var will hold June 21 2026
	birthDay := time.Date(current.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, time.UTC)

	// handle pple born on leap yr
	if dob.Month() == time.February && dob.Day() == 29 && birthDay.Equal(time.Date(current.Year(),
		time.March, 1, 0, 0, 0, 0, time.UTC)) {
		birthDay = time.Date(current.Year(), time.February, 28, 0, 0, 0, 0, time.UTC)
	}

	if current.Before(birthDay) {
		age--
	}
	return age
}

// amountToCents converts the amount from float64 to integer cents(int64)
func amountToCents(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 || amount > 9e15 {
		return 0, domain.ErrInvalidAmount
	}

	// convert amount to cents
	cent := amount * 100

	// check if the amount can be represented as a whole number of cents
	if math.Abs(cent-math.Round(cent)) > 0.000001 {
		return 0, domain.ErrInvalidAmount
	}

	//round to the nearest integer then return the value
	return int64(math.Round(cent)), nil
}
