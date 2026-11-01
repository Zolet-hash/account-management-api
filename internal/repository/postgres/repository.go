package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

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

// safe transaction transfers with idempotency support
func (r *Repository) Transfer(ctx context.Context, id uuid.UUID, typ string, amount int64, key string) (model.Transaction, error) {

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Transaction{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var existing model.Transaction
	var oldAmount string

	err = tx.QueryRow(ctx, `SELECT id, account_id, type,amount::text, created_at FROM transactions 
	WHERE account_id=$1 AND idempotency_key=$2`, id, key,
	).Scan(&existing.ID, &existing.AccountID, &existing.Type, &oldAmount, &existing.CreatedAt)

	if err == nil {
		old, err := tocent(oldAmount)
		if err != nil {
			return existing, err
		}

		if existing.Type != typ || old != amount {
			return existing, domain.ErrIdempotencyConflict
		}

		existing.AmountCents = old
		existing.Amount = float64(old) / 100
		return existing, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return existing, err
	}

	var balanceText, limitText string
	err = tx.QueryRow(ctx, `SELECT account.balance::text, exchange.maximum_transfer_amount::text 
	FROM accounts account JOIN exchanges exchange ON exhange.id=account.exchange_id WHERE account.id=$1 FOR UPDATE OF account`, id,
	).Scan(&balanceText, &limitText)

	if err != nil {
		return existing, mapErr(err)
	}
	// A competing request may have committed the same key while we waited for the row lock.
	err = tx.QueryRow(ctx, `SELECT id, account_id, type,amount::text,created_at FROM transactions
		 WHERE account_id=$1 AND idempotency_key=$2`, id, key,
	).Scan(&existing.ID, &existing.AccountID, &existing.Type, &oldAmount, &existing.CreatedAt)

	if err == nil {
		old, parseErr := tocent(oldAmount)
		if parseErr != nil {
			return existing, parseErr
		}

		if existing.Type != typ || old != amount {
			return existing, domain.ErrIdempotencyConflict
		}

		if err = tx.Commit(ctx); err != nil {
			return existing, err
		}

		existing.AmountCents = old
		existing.Amount = float64(old) / 100
		return existing, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return existing, err
	}
	bal, err := tocent(balanceText)

	if err != nil {
		return existing, err
	}
	limit, err := tocent(limitText)

	if err != nil {
		return existing, err
	}

	if amount <= 0 {
		return existing, domain.ErrInvalidAmount
	}

	if amount > limit {
		return existing, domain.ErrTransferLimitExceeded
	}
	delta := amount

	if typ == "withdrawal" {
		delta = -amount
		if amount > bal {
			return existing, domain.ErrInsufficientBalance
		}
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET balance=balance+$2::numeric,updated_at=now() WHERE id=$1`, id, money(delta))
	if err != nil {
		return existing, err
	}

	existing = model.Transaction{
		ID:          uuid.New(),
		AccountID:   id,
		Type:        typ,
		AmountCents: amount,
		Amount:      float64(amount) / 100,
	}

	err = tx.QueryRow(ctx, `INSERT INTO transactions(id, account_id, type, amount, idempotency_key) 
			VALUES($1,$2,$3,$4::numeric,$5) 
			RETURNING created_at`, existing.ID, id, typ, money(amount), key,
	).Scan(&existing.CreatedAt)

	if err != nil {
		return existing, mapErr(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return existing, err
	}
	return existing, nil
}

// Balances calculates users balance
func (r *Repository) Balances(ctx context.Context, id uuid.UUID) (model.UserBalances, error) {

	balances := model.UserBalances{
		UserID:   id,
		Balances: []model.Balance{},
	}

	rows, err := r.DB.Query(ctx, `SELECT exchange.id, exchange.name, account.balance::text 
			FROM accounts account JOIN exchanges exchange ON exchange.id=account.exchange_id 
			WHERE account.user_id=$1 AND account.balance>0 ORDER BY exchange.name`, id)

	if err != nil {
		return balances, err
	}
	defer rows.Close()

	var totalCents int64
	for rows.Next() {
		var balance model.Balance
		var amount string

		if err = rows.Scan(&balance.ExchangeID, &balance.ExchangeName, &amount); err != nil {
			return balances, err
		}
		centAmount, err := tocent(amount)

		if err != nil {
			return balances, err
		}
		balance.Amount = float64(centAmount) / 100
		totalCents += centAmount
		balances.Balances = append(balances.Balances, balance)
	}
	balances.TotalBalance = float64(totalCents) / 100
	return balances, rows.Err()
}

// Volumes retrieves daily transaction volumes grouped by exchange
// within the specified date range. It includes days with no transactions
// and returns each exchange's total volume for each day.
func (r *Repository) Volumes(ctx context.Context, start, end time.Time) ([]model.DailyVolume, error) {

	days := int(end.Sub(start).Hours()/24) + 1
	out := make([]model.DailyVolume, days)

	for i := range out {
		out[i] = model.DailyVolume{
			Date:            start.AddDate(0, 0, i).Format("2006-01-02"),
			ExchangeVolumes: []model.ExchangeVolume{}}
	}

	rows, err := r.DB.Query(ctx, `SELECT (trans.created_at AT TIME ZONE 'UTC')::date::text,
	 		exchage.id, exchange.name, sum(trans.amount)::text FROM transactions trans 
	 		JOIN accounts account ON account.id=trans.account_id JOIN exchanges exchange
			ON exchange.id=account.exchange_id WHERE trans.created_at >= ($1::date::timestamp AT TIME ZONE 'UTC') 
			AND trans.created_at < (($2::date+interval '1 day')::timestamp AT TIME ZONE 'UTC') 
			GROUP BY 1,2,3 ORDER BY 1,3`, start.UTC().Format("2006-01-02"), end.UTC().Format("2006-01-02"))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var date string
		var exchangeVol model.ExchangeVolume
		var amount string

		if err = rows.Scan(&date, &exchangeVol.ExchangeID, &exchangeVol.ExchangeName, &amount); err != nil {
			return nil, err
		}

		centAmount, err := tocent(amount)
		if err != nil {
			return nil, err
		}
		exchangeVol.Volume = float64(centAmount) / 100

		format := int(mustDate(date).Sub(start).Hours() / 24)
		if format >= 0 && format < len(out) {
			out[format].ExchangeVolumes = append(out[format].ExchangeVolumes, exchangeVol)
		}
	}
	return out, rows.Err()
}

// mustDate converts a date string into a time.Time value using the YYYY-MM-DD format.
func mustDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", strings.TrimSpace(s))
	return t
}
