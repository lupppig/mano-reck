package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// InTransaction executes an application operation in one PostgreSQL
// transaction. The operation owns no commit and can only request rollback by
// returning an error.
func (database *Database) InTransaction(
	ctx context.Context,
	options pgx.TxOptions,
	operation func(DBTX) error,
) error {
	if operation == nil {
		return fmt.Errorf("database transaction operation must not be nil")
	}
	pool, err := database.Pool()
	if err != nil {
		return err
	}

	transaction, err := pool.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin database transaction: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	if err := operation(transaction); err != nil {
		rollbackError := transaction.Rollback(ctx)
		if rollbackError != nil && !errors.Is(rollbackError, pgx.ErrTxClosed) {
			return errors.Join(err, fmt.Errorf("rollback database transaction: %w", rollbackError))
		}
		return err
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit database transaction: %w", err)
	}
	return nil
}
