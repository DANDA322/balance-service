package models

import "time"

type Wallet struct {
	ID              int       `json:"id" db:"id"`
	Owner           int       `json:"owner" db:"owner_id"`
	Balance         float64   `json:"balance" db:"balance"`
	ReservedBalance float64   `json:"reserved_balance" db:"reserved_balance"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
