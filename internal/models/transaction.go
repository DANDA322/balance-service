package models

import "time"

type Transaction struct {
	IdempotenceKey int     `json:"idempotence_key"`
	Amount         float64 `json:"amount"`
	Comment        string  `json:"comment"`
}

type TransferTransaction struct {
	IdempotenceKey int     `json:"idempotence_key"`
	Target         int     `json:"target"`
	Amount         float64 `json:"amount"`
	Comment        string  `json:"comment"`
}

type ReserveTransaction struct {
	AccountID int     `json:"account_id"`
	ServiceID int     `json:"service_id"`
	OrderID   int     `json:"order_id"`
	Amount    float64 `json:"amount"`
}

type TransactionFullInfo struct {
	ID             int       `json:"id" db:"id"`
	WalletID       int       `json:"wallet_id" db:"wallet_id"`
	Amount         float64   `json:"amount" db:"amount"`
	TargetWalletID *int      `json:"target_wallet_id" db:"target_wallet_id"`
	ServiceID      *int      `json:"service_id" db:"service_id"`
	Comment        string    `json:"comment" db:"comment"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
}
