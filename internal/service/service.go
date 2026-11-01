package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/domain"
	"github.com/Zolet-hash/account-management-api/internal/model"
	"github.com/google/uuid"
)

// group interface togath
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

type Service struct {
	store Store
	now   func() time.Time
}

func New(s Store) *Service {

	return &Service{store: s, now: time.Now}
}

func (s *Service) CreateUser(ctx context.Context, username string, dob time.Time, document string) (model.User, error) {
	if username == "" || document == "" || dob.IsZero() || dob.After(s.now()) {
		return model.User{}, domain.ErrInvalidRequest
	}
	return s.store.CreateUser(ctx, model.User{ID: uuid.New(), Username: username, DateOfBirth: dob, DocumentNumber: document})
}

// create exchanges account
func (s *Service) CreateAccount(ctx context.Context, userID, exchangeID uuid.UUID) (model.Account, error) {
	user, err := s.store.User(ctx, userID)
	if err != nil {
		return model.Account{}, err
	}
	exchange, err := s.store.Exchange(ctx, exchangeID)
	if err != nil {
		return model.Account{}, err
	}
	if calculateAge(user.DateOfBirth, s.now().UTC()) < exchange.MinimumAge {
		return model.Account{}, domain.ErrMinimumAgeNotMet
	}
	return s.store.CreateAccount(ctx, model.Account{ID: uuid.New(), UserID: userID, ExchangeID: exchangeID})
}

func (s *Service) Transfer(ctx context.Context, id uuid.UUID, typ string, amount float64, key string) (model.Transaction, error) {
	if typ != "deposit" && typ != "withdrawal" {
		return model.Transaction{}, domain.ErrInvalidRequest
	}
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return model.Transaction{}, domain.ErrInvalidRequest
	}
	cent, err := amountToCents(amount)
	if err != nil {
		return model.Transaction{}, err
	}
	_, exchange, err := s.store.Account(ctx, id)
	if err != nil {
		return model.Transaction{}, err
	}
	if cent > exchange.MaximumTransferCents {
		return model.Transaction{}, domain.ErrTransferLimitExceeded
	}
	return s.store.Transfer(ctx, id, typ, cent, key)
}

func (s *Service) Balances(ctx context.Context, id uuid.UUID) (model.UserBalances, error) {
	if _, err := s.store.User(ctx, id); err != nil {
		return model.UserBalances{}, err
	}
	return s.store.Balances(ctx, id)
}

func (s *Service) Volumes(ctx context.Context, start, end time.Time) ([]model.DailyVolume, error) {
	start = start.UTC()
	end = end.UTC()
	if start.After(end) || end.Sub(start) > 90*24*time.Hour {
		return nil, fmt.Errorf("%w: range must be 0-90 days", domain.ErrInvalidRequest)
	}
	return s.store.Volumes(ctx, start, end)
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
