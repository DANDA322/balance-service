package pgstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const dateTimeFmt = "2006-01-02 15:04:05"

type DB struct {
	log *logrus.Logger
	db  *sqlx.DB
	dsn string
}

func GetPGStore(ctx context.Context, log *logrus.Logger, dsn string) (*DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &DB{
		log: log,
		db:  db,
		dsn: dsn,
	}, nil
}

func (db *DB) GetWallet(ctx context.Context, accountID int) (*models.Wallet, error) {
	query := `
	SELECT id, balance, reserved_balance, created_at, updated_at
	FROM wallet
	WHERE owner_id = $1`
	var wallet models.Wallet
	if err := db.db.GetContext(ctx, &wallet, query, accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrWalletNotFound
		}
		return nil, fmt.Errorf("err executing [GetWallet]: %w", err)
	}
	return &wallet, nil
}

func (db *DB) UpsertDepositToWallet(ctx context.Context, ownerID int, transaction models.Transaction) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("err deposit money the wallet: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			db.log.Error("err rolling back deposit transaction")
		}
	}()
	query := `
	INSERT INTO wallet (owner_id, balance, reserved_balance, created_at, updated_at)
	VALUES ($1, $2, 0, $3, $3)
	ON CONFLICT (owner_id) DO UPDATE SET balance = wallet.balance + excluded.balance,
										updated_at = excluded.updated_at`
	if _, err = tx.ExecContext(ctx, query, ownerID, transaction.Amount, time.Now().UTC().Format(dateTimeFmt)); err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	wallet, err := db.checkBalance(ctx, tx, ownerID, 0)
	if err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	if err = db.insertTransaction(ctx, tx, wallet.ID, nil, transaction.Amount, nil, transaction.Comment); err != nil {
		return fmt.Errorf("err executing [UpsertDepositToWallet]: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("err committing the transaction: %w", err)
	}
	return nil
}

func (db *DB) WithdrawMoneyFromWallet(ctx context.Context, ownerID int, transaction models.Transaction) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("err withdraw money the wallet: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			db.log.Error("err rolling back withdraw transaction")
		}
	}()
	wallet, err := db.checkBalance(ctx, tx, ownerID, transaction.Amount)
	if err != nil {
		return err
	}
	query := `
	UPDATE wallet 
	SET balance = balance - $1,
	updated_at = $3
	WHERE owner_id = $2`
	result, err := tx.ExecContext(ctx, query, transaction.Amount, ownerID, time.Now().UTC().Format(dateTimeFmt))
	if err != nil {
		return fmt.Errorf("err executing [WithdrawMoneyFromWallet]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	transaction.Amount *= -1
	if err = db.insertTransaction(ctx, tx, wallet.ID, nil, transaction.Amount, nil, transaction.Comment); err != nil {
		return fmt.Errorf("err executing [WithdrawMoneyFromWallet]: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("err committing the transaction: %w", err)
	}
	return nil
}

func (db *DB) TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("err transfer money from <%d> to <%d>: %w", accountID, transaction.Target, err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			db.log.Error("err rolling back withdraw transaction")
		}
	}()
	wallet, err := db.checkBalance(ctx, tx, accountID, transaction.Amount)
	if err != nil {
		return err
	}
	err = db.withdrawMoney(ctx, tx, wallet.ID, transaction.Amount)
	if err != nil {
		return err
	}
	targetWallet, err := db.checkBalance(ctx, tx, transaction.Target, 0)
	if err != nil {
		return err
	}
	err = db.depositMoney(ctx, tx, targetWallet.ID, transaction.Amount)
	if err != nil {
		return err
	}
	err = db.insertTransaction(ctx, tx, wallet.ID, &targetWallet.ID, transaction.Amount, nil, transaction.Comment)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("err committing the transaction: %w", err)
	}
	return nil
}

func (db *DB) checkBalance(ctx context.Context, tx *sql.Tx, ownerID int, amount float64) (*models.Wallet, error) {
	query := `
	SELECT id, balance
	FROM wallet
	WHERE owner_id = $1
	FOR UPDATE`
	row := tx.QueryRowContext(ctx, query, ownerID)
	var wallet models.Wallet
	if err := row.Scan(&wallet.ID, &wallet.Balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrWalletNotFound
		}
		return nil, fmt.Errorf("err checking balance: %w", err)
	}
	if amount == 0 {
		return &wallet, nil
	}
	if wallet.Balance-amount < 0 {
		return nil, models.ErrNotEnoughMoney
	}
	return &wallet, nil
}

func (db *DB) insertTransaction(ctx context.Context, tx *sql.Tx, walletID int, targetWalletID *int, amount float64,
	serviceID *int, comment string) error {
	query := `
	INSERT INTO transaction (wallet_id, amount, target_wallet_id, service_id, comment, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.ExecContext(ctx, query, walletID, amount, targetWalletID, serviceID, comment, time.Now().UTC().Format(dateTimeFmt))
	if err != nil {
		return fmt.Errorf("err executing [insertTransaction]: %w", err)
	}
	return nil
}

func (db *DB) withdrawMoney(ctx context.Context, tx *sql.Tx, walletID int, amount float64) error {
	query := `
	UPDATE wallet 
	SET balance = balance - $1,
	updated_at = $3
	WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, walletID, time.Now().UTC().Format(dateTimeFmt))
	if err != nil {
		return fmt.Errorf("err executing [withdrawMoney]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}

func (db *DB) depositMoney(ctx context.Context, tx *sql.Tx, walletID int, amount float64) error {
	query := `
	UPDATE wallet 
	SET balance = balance + $1,
	updated_at = $3
	WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, walletID, time.Now().UTC().Format(dateTimeFmt))
	if err != nil {
		return fmt.Errorf("err executing [depositMoney]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}
