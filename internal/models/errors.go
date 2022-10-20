package models

import "errors"

var (
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrOrderNotFound          = errors.New("reserved order not found")
	ErrServiceNotFound        = errors.New("service not found")
	ErrNotEnoughMoney         = errors.New("not enough money on the balance")
	ErrNotEnoughReservedMoney = errors.New("not enough reserved money on the balance")
)
