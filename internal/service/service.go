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
