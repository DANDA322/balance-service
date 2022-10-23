package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

const (
	dateTimeLayout = "2006-01-02 15:04:05"
	retries        = 3
)

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

func (db *DB) Migrate(direction migrate.MigrationDirection) error {
	conn, err := sql.Open("pgx", db.dsn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			db.log.Errorf("err closing migrations connections")
		}
	}()
	asserDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, err := migrations.ReadDir(path)
			if err != nil {
				return nil, err
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}
			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: asserDir,
		Dir:      "migrations",
	}
	_, err = migrate.Exec(conn, "postgres", asset, direction)
	return err
}

func (db *DB) GetWallet(ctx context.Context, accountID int) (*models.Wallet, error) {
	query := `
	SELECT id, balance, reserved_balance, created_at, updated_at
	FROM wallet
	WHERE owner_id = $1`
	var wallet models.Wallet
	var err error
	for i := 0; i < retries; i++ {
		if err = db.db.GetContext(ctx, &wallet, query, accountID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, models.ErrWalletNotFound
			}
			continue
		}
		return &wallet, nil
	}
	return nil, err
}

func (db *DB) UpsertDepositToWallet(ctx context.Context, ownerID int, transaction models.Transaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		query := `
	INSERT INTO wallet (owner_id, balance, reserved_balance, created_at, updated_at)
	VALUES ($1, $2, 0, $3, $3)
	ON CONFLICT (owner_id) DO UPDATE SET balance = wallet.balance + excluded.balance,
										updated_at = excluded.updated_at`
		if _, err = tx.ExecContext(ctx, query, ownerID, transaction.Amount, time.Now().UTC().Format(dateTimeLayout)); err != nil {
			_ = tx.Rollback()
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkBalance(ctx, tx, ownerID, 0)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		if err = db.insertTransaction(ctx, tx, transaction.IdempotenceKey, wallet.ID, nil, transaction.Amount,
			nil, transaction.Comment); err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) WithdrawMoneyFromWallet(ctx context.Context, ownerID int, transaction models.Transaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkBalance(ctx, tx, ownerID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		query := `
	UPDATE wallet 
	SET balance = balance - $1,
	updated_at = $3
	WHERE owner_id = $2`
		var result sql.Result
		result, err = tx.ExecContext(ctx, query, transaction.Amount, ownerID, time.Now().UTC().Format(dateTimeLayout))
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		if count, _ := result.RowsAffected(); count == 0 {
			return models.ErrWalletNotFound
		}
		transaction.Amount *= -1
		if err = db.insertTransaction(ctx, tx, transaction.IdempotenceKey, wallet.ID, nil, transaction.Amount,
			nil, transaction.Comment); err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkBalance(ctx, tx, accountID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.withdrawMoney(ctx, tx, wallet.ID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.depositMoney(ctx, tx, transaction.Target, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.insertTransaction(ctx, tx, transaction.IdempotenceKey, wallet.ID, &transaction.Target, transaction.Amount,
			nil, transaction.Comment)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) ReserveMoneyFromWallet(ctx context.Context, transaction models.ReserveTransaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkBalance(ctx, tx, transaction.AccountID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.withdrawMoney(ctx, tx, wallet.ID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.reserveMoney(ctx, tx, wallet.ID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.insertReservedFunds(ctx, tx, transaction.AccountID, transaction)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) ApplyReservedMoney(ctx context.Context, transaction models.ReserveTransaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkReservedBalance(ctx, tx, transaction.AccountID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.withdrawReservedMoney(ctx, tx, wallet.ID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.updateOrderStatus(ctx, tx, transaction.AccountID, "Completed", transaction)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		var serviceTitle string
		serviceTitle, err = db.getServiceTitle(ctx, tx, transaction.ServiceID)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.insertTransaction(ctx, tx, transaction.OrderID, wallet.ID, nil, transaction.Amount,
			&transaction.ServiceID, serviceTitle)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) CancelReserve(ctx context.Context, transaction models.ReserveTransaction) error {
	var err error
	for i := 0; i < retries; i++ {
		var tx *sql.Tx
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}
		var wallet *models.Wallet
		wallet, err = db.checkReservedBalance(ctx, tx, transaction.AccountID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.withdrawReservedMoney(ctx, tx, wallet.ID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.depositMoney(ctx, tx, transaction.AccountID, transaction.Amount)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = db.updateOrderStatus(ctx, tx, transaction.AccountID, "Cancelled", transaction)
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()
			continue
		}
		return nil
	}
	return err
}

func (db *DB) GetWalletTransactions(ctx context.Context, accountID int,
	queryParams *models.TransactionsQueryParams) ([]models.TransactionFullInfo, error) {
	var err error
	for i := 0; i < retries; i++ {
		var transactions []models.TransactionFullInfo
		err = func([]models.TransactionFullInfo) error {
			var wallet *models.Wallet
			wallet, err = db.GetWallet(ctx, accountID)
			if err != nil {
				return err
			}
			query := db.queryBuilder(queryParams.Sorting, queryParams.Descending)
			var rows *sqlx.Rows
			rows, err = db.db.QueryxContext(ctx, query, wallet.ID, queryParams.From, queryParams.To,
				queryParams.Limit, queryParams.Offset)
			if err != nil {
				return err
			}
			defer func() {
				if err = rows.Close(); err != nil {
					db.log.Warnf("err closing rows: %v", err)
				}
			}()
			var otherTransaction models.TransactionFullInfo
			for rows.Next() {
				if err = rows.StructScan(&otherTransaction); err != nil {
					continue
				}
				transactions = append(transactions, otherTransaction)
			}
			return nil
		}(transactions)
		if err != nil {
			continue
		}
		return transactions, nil
	}
	return nil, err
}

func (db *DB) GetReport(ctx context.Context, month time.Time) (map[string]float64, error) {
	query := `
	SELECT title, amount
	FROM reserved_funds
	INNER JOIN services s on s.id = reserved_funds.service_id
	WHERE status = $1 AND 
	    updated_at BETWEEN $2 AND $3`
	status := "Completed"
	var err error
	for i := 0; i < retries; i++ {
		services := make(map[string]float64)
		err = func(map[string]float64) error {
			var rows *sqlx.Rows
			rows, err = db.db.QueryxContext(ctx, query, status, month, month.AddDate(0, 1, 0))
			if err != nil {
				return err
			}
			defer func() {
				if err = rows.Close(); err != nil {
					db.log.Warnf("err closing rows: %v", err)
				}
			}()
			var service models.Service
			for rows.Next() {
				if err = rows.StructScan(&service); err != nil {
					return err
				}
				services[service.Title] += service.Amount
			}
			return nil
		}(services)
		if err != nil {
			continue
		}
		return services, nil
	}
	return nil, err
}

func (db *DB) reserveMoney(ctx context.Context, tx *sql.Tx, walletID int, amount float64) error {
	query := `
	UPDATE wallet 
	SET reserved_balance = reserved_balance + $1,
	updated_at = $3
	WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, walletID, time.Now().UTC().Format(dateTimeLayout))
	if err != nil {
		return fmt.Errorf("err executing [addReserve]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}

func (db *DB) insertReservedFunds(ctx context.Context, tx *sql.Tx, accountID int, transaction models.ReserveTransaction) error {
	query := `
	INSERT INTO reserved_funds (order_id, owner_id, service_id, amount, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $6)`
	status := "Active"
	_, err := tx.ExecContext(ctx, query, transaction.OrderID, accountID, transaction.ServiceID, transaction.Amount,
		status, time.Now().UTC().Format(dateTimeLayout))
	if err != nil {
		return fmt.Errorf("err executing [insertReservedFunds]: %w", err)
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

func (db *DB) updateOrderStatus(ctx context.Context, tx *sql.Tx, ownerID int, status string,
	transaction models.ReserveTransaction) error {
	query := `
	UPDATE reserved_funds 
	SET status = $1,
	updated_at = $2
	WHERE owner_id = $3 AND 
	      service_id = $4 AND
	      order_id = $5 AND 
	      amount = $6 AND
	      status = $7`
	result, err := tx.ExecContext(ctx, query, status, time.Now().UTC().Format(dateTimeLayout), ownerID, transaction.ServiceID,
		transaction.OrderID, transaction.Amount, "Active")
	if err != nil {
		return fmt.Errorf("err executing [updateOrderStatus]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrOrderNotFound
	}
	return nil
}

func (db *DB) checkReservedBalance(ctx context.Context, tx *sql.Tx, ownerID int, amount float64) (*models.Wallet, error) {
	query := `
	SELECT id, reserved_balance
	FROM wallet
	WHERE owner_id = $1
	FOR UPDATE`
	row := tx.QueryRowContext(ctx, query, ownerID)
	var wallet models.Wallet
	if err := row.Scan(&wallet.ID, &wallet.ReservedBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrWalletNotFound
		}
		return nil, fmt.Errorf("err checking reserved balance: %w", err)
	}
	if wallet.ReservedBalance-amount < 0 {
		return nil, models.ErrNotEnoughReservedMoney
	}
	return &wallet, nil
}

func (db *DB) insertTransaction(ctx context.Context, tx *sql.Tx, idempotenceKey, walletID int, targetOwnerID *int, amount float64,
	serviceID *int, comment string) error {
	var query string
	var err error
	if targetOwnerID == nil {
		query = `
		INSERT INTO transaction (idempotence_key, wallet_id, amount, target_wallet_id, service_id, comment, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	} else {
		query = `
	INSERT INTO transaction (idempotence_key, wallet_id, amount, target_wallet_id, service_id, comment, timestamp)
	VALUES ($1, $2, $3, (SELECT id FROM wallet where owner_id = $4), $5, $6, $7)`
	}
	_, err = tx.ExecContext(ctx, query, idempotenceKey, walletID, amount, targetOwnerID, serviceID, comment,
		time.Now().UTC().Format(dateTimeLayout))
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
	result, err := tx.ExecContext(ctx, query, amount, walletID, time.Now().UTC().Format(dateTimeLayout))
	if err != nil {
		return fmt.Errorf("err executing [withdrawMoney]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}

func (db *DB) withdrawReservedMoney(ctx context.Context, tx *sql.Tx, walletID int, amount float64) error {
	query := `
	UPDATE wallet 
	SET reserved_balance = reserved_balance - $1,
	updated_at = $3
	WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, walletID, time.Now().UTC().Format(dateTimeLayout))
	if err != nil {
		return fmt.Errorf("err executing [withdrawReservedMoney]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}

func (db *DB) depositMoney(ctx context.Context, tx *sql.Tx, ownerID int, amount float64) error {
	query := `
	UPDATE wallet 
	SET balance = balance + $1,
	updated_at = $3
	WHERE wallet.id in (SELECT id from wallet WHERE wallet.owner_id = $2)`
	result, err := tx.ExecContext(ctx, query, amount, ownerID, time.Now().UTC().Format(dateTimeLayout))
	if err != nil {
		return fmt.Errorf("err executing [depositMoney]: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return models.ErrWalletNotFound
	}
	return nil
}

func (db *DB) getServiceTitle(ctx context.Context, tx *sql.Tx, serviceID int) (string, error) {
	query := `
	SELECT title
	FROM services
	WHERE id = $1`
	row := tx.QueryRowContext(ctx, query, serviceID)
	var title string
	if err := row.Scan(&title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", models.ErrServiceNotFound
		}
		return "", fmt.Errorf("err getting service title: %w", err)
	}
	return title, nil
}

func (db *DB) queryBuilder(sorting, descending string) string {
	query := `SELECT id, wallet_id, amount, target_wallet_id, service_id, comment, timestamp
	FROM transaction
	WHERE (wallet_id = $1 OR target_wallet_id = $1)
	AND timestamp BETWEEN $2 AND $3`

	switch sorting {
	case "date":
		query += " ORDER BY timestamp"
	case "amount":
		query += " ORDER BY amount"
	default:
		query += " ORDER BY timestamp"
	}
	switch descending {
	case "true":
		query += " DESC"
	case "false":
		query += " ASC"
	default:
		query += " DESC"
	}
	query += " LIMIT $4 OFFSET $5"
	return query
}
