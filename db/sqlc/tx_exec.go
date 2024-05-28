package db

import (
	"context"
	"fmt"
)

// execTx executes a function within a database transaction
// method starts with lower case to not be exported so external packages can't call it directly
// we will provide an exported function for each specific transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)

}
