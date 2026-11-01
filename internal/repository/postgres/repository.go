package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"uuid"

	"github.com/Zolet-hash/account-management-api/internal/domain"
	"github.com/Zolet-hash/account-management-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func money(cent int64) string {
	if cent < 0 {
		amount := abs(cent)
		return fmt.Sprintf("-%d.%02d", amount/100, amount%100)
	}
	return fmt.Sprintf("%d.%02d", cent/100, cent%100)
}

// abs calculates absolute value of an int
func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

// converts amountfrom string to cent
func tocent(value string) (int64, error) {

	// cleanup the string
	value = strings.TrimSpace(value) // remove unecessary spaces
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "-"), "+")

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid numeric amount")
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric amount: %w", err)
	}

	fraction := "00"
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, fmt.Errorf("numeric amount has more than two decimal places")
		}
		fraction = parts[1] + strings.Repeat("0", 2-len(parts[1]))
	}
	// convert fraction into integer
	minor, err := strconv.ParseInt(fraction, 10, 64)
	// check for int overflow
	if err != nil || whole > (math.MaxInt64-minor)/100 {
		return 0, fmt.Errorf("numeric amount overflows cents")
	}

	result := whole*100 + minor
	if negative {
		result = -result
	}

	return result, nil
}

// mapErr maps postgres errors
func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		if pgerr.Code == "23505" {
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgerr.ConstraintName)
		}
		if pgerr.Code == "23503" {
			return domain.ErrNotFound
		}
		if pgerr.Code == "23514" {
			return domain.ErrInvalidRequest
		}
	}
	return err
}

// CreateUser inserts user data into the database
func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {

	err := r.DB.QueryRow(ctx, `INSERT INTO users(id, username, date_of_birth, document_number) 
	VALUES($1,$2,$3,$4) RETURNING created_at`, user.ID, user.Username, user.DateOfBirth, user.DocumentNumber,
	).Scan(&user.CreatedAt)

	return user, mapErr(err)
}

func (r *Repository) User(ctx context.Context, id uuid.UUID) (model.User, error) {

	var user model.User
	err := r.DB.QueryRow(ctx, `SELECT id, username, date_of_birth, document_number,
								created_at FROM users WHERE id=$1`, id,
	).Scan(&user.ID, &user.Username, &user.DateOfBirth, &user.DocumentNumber, &user.CreatedAt)
	return user, mapErr(err)
}

func (r *Repository) Exchange(ctx context.Context, id uuid.UUID) (model.Exchange, error) {

	var exchange model.Exchange
	var amount string

	err := r.DB.QueryRow(ctx, `SELECT id, name,minimum_age,
		maximum_transfer_amount::text FROM exchanges WHERE id=$1`, id,
	).Scan(&exchange.ID, &exchange.Name, &exchange.MinimumAge, &amount)

	if err != nil {
		return exchange, mapErr(err)
	}

	exchange.MaximumTransferCents, err = tocent(amount)
	exchange.MaximumTransferAmount = float64(exchange.MaximumTransferCents) / 100
	return exchange, err
}

// CreateAccount inserts accountdata into the database
func (r *Repository) CreateAccount(ctx context.Context, account model.Account) (model.Account, error) {
	err := r.DB.QueryRow(ctx, `INSERT INTO accounts(id, user_id,exchange_id) 
	VALUES($1,$2,$3) 
	RETURNING created_at`, account.ID, account.UserID, account.ExchangeID).Scan(&account.CreatedAt)
	return account, mapErr(err)
}

func (r *Repository) Account(ctx context.Context, id uuid.UUID) (model.Account, model.Exchange, error) {

	var account model.Account
	var exchange model.Exchange
	var amount, limit string

	err := r.DB.QueryRow(ctx, `SELECT account.id, account.user_id, account.exchange_id, 
	account.balance::text, account.created_at, exchange.name, exchange.minimum_age, 
	exchange.maximum_transfer_amount::text FROM accounts account JOIN exchanges exchange ON exchange.id=account.exchange_id WHERE account.id=$1`, id,
	).Scan(&account.ID, &account.UserID, &account.ExchangeID, &amount, &account.CreatedAt, &exchange.Name, &exchange.MinimumAge, &limit)
	
	if err != nil {
		return account, exchange, mapErr(err)
	}

	account.BalanceCents, err = tocent(amount)
	if err != nil {
		return account, exchange, err
	}

	account.Balance = float64(account.BalanceCents) / 100
	exchange.ID = account.ExchangeID

	exchange.MaximumTransferCents, err = tocent(limit)

	if err != nil {
		return account, exchange, err
	}
	exchange.MaximumTransferAmount = float64(exchange.MaximumTransferCents) / 100
	return account, exchange, err
}
