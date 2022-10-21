package internal

import (
	"context"
	"fmt"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/sirupsen/logrus"
)

type Database interface {
	GetWallet(ctx context.Context, accountID int) (*models.Wallet, error)
	UpsertDepositToWallet(ctx context.Context, accountID int, transaction models.Transaction) error
	WithdrawMoneyFromWallet(ctx context.Context, ownerID int, transaction models.Transaction) error
	TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error
	ReserveMoneyFromWallet(ctx context.Context, accountID int, transaction models.ReserveTransaction) error
	ApplyReservedMoney(ctx context.Context, accountID int, transaction models.ReserveTransaction) error
	GetWalletTransactions(ctx context.Context, accountID int,
		queryParams *models.TransactionsQueryParams) ([]models.TransactionFullInfo, error)
	CancelReserve(ctx context.Context, accountID int, transaction models.ReserveTransaction) error
}

type App struct {
	log *logrus.Logger
	db  Database
}

func NewApp(log *logrus.Logger, db Database) *App {
	return &App{
		log: log,
		db:  db,
	}
}

func (a *App) AddDeposit(ctx context.Context, accountID int, transaction models.Transaction) error {
	if err := a.db.UpsertDepositToWallet(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to upsert deposit: %w", err)
	}
	return nil
}

func (a *App) WithdrawMoney(ctx context.Context, accountID int, transaction models.Transaction) error {
	if err := a.db.WithdrawMoneyFromWallet(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to withdraw money: %w", err)
	}
	return nil
}

func (a *App) TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error {
	if err := a.db.TransferMoney(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to transfer mony: %w", err)
	}
	return nil
}

func (a *App) GetBalance(ctx context.Context, accountID int) (float64, error) {
	wallet, err := a.db.GetWallet(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("unable to get balance: %w", err)
	}
	return wallet.Balance, nil
}

func (a *App) ReserveMoney(ctx context.Context, accountID int, transaction models.ReserveTransaction) error {
	if err := a.db.ReserveMoneyFromWallet(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to reserve money: %w", err)
	}
	return nil
}

func (a *App) ApplyReservedMoney(ctx context.Context, accountID int, transaction models.ReserveTransaction) error {
	if err := a.db.ApplyReservedMoney(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to recognize money: %w", err)
	}
	return nil
}

func (a *App) GetWalletTransaction(ctx context.Context, accountID int,
	queryParams *models.TransactionsQueryParams) ([]models.TransactionFullInfo, error) {
	transactions, err := a.db.GetWalletTransactions(ctx, accountID, queryParams)
	if err != nil {
		return nil, fmt.Errorf("unable to get transactions: %w", err)
	}
	return transactions, nil
}

func (a *App) CancelReserve(ctx context.Context, accountID int, transaction models.ReserveTransaction) error {
	if err := a.db.CancelReserve(ctx, accountID, transaction); err != nil {
		return fmt.Errorf("unable to cancel reserve")
	}
	return nil
}
