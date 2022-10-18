package models

import "errors"

var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrNotEnoughMoney = errors.New("not enough money on the balance")
)
