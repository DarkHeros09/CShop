package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	FinishedPurchaseTx(ctx context.Context, arg FinishedPurchaseTxParams) (*FinishedPurchaseTxResult, error)
	DeleteShopOrderItemTx(ctx context.Context, arg DeleteShopOrderItemTxParams) error
	SignUpTx(ctx context.Context, arg SignUpTxParams) (*SignUpTxResult, error)
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
