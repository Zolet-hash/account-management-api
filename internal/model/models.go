package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	DateOfBirth    time.Time `json:"dateOfBirth"`
	DocumentNumber string    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
}
type Exchange struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	MinimumAge            int       `json:"minimumAge"`
	MaximumTransferCents  int64     `json:"-"`
	MaximumTransferAmount float64   `json:"maximumTransferAmount"`
}
type Account struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	ExchangeID   uuid.UUID `json:"exchangeId"`
	BalanceCents int64     `json:"-"`
	Balance      float64   `json:"balance"`
	CreatedAt    time.Time `json:"createdAt"`
}
type Transaction struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"accountId"`
	Type        string    `json:"type"`
	AmountCents int64     `json:"-"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"createdAt"`
}
type Balance struct {
	ExchangeID   uuid.UUID `json:"exchangeId"`
	ExchangeName string    `json:"exchangeName"`
	Amount       float64   `json:"amount"`
}
type UserBalances struct {
	UserID       uuid.UUID `json:"userId"`
	Balances     []Balance `json:"balances"`
	TotalBalance float64   `json:"totalBalance"`
}
type ExchangeVolume struct {
	ExchangeID   uuid.UUID `json:"exchangeId"`
	ExchangeName string    `json:"exchangeName"`
	Volume       float64   `json:"volume"`
}
type DailyVolume struct {
	Date            string           `json:"date"`
	ExchangeVolumes []ExchangeVolume `json:"exchangeVolumes"`
}
