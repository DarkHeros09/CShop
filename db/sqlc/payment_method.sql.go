// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: payment_method.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createPaymentMethod = `-- name: CreatePaymentMethod :one
INSERT INTO "payment_method" (
  user_id,
  payment_type_id,
  provider
) VALUES (
  $1, $2, $3
)
RETURNING id, user_id, payment_type_id, provider, is_default
`

type CreatePaymentMethodParams struct {
	UserID        int64  `json:"user_id"`
	PaymentTypeID int64  `json:"payment_type_id"`
	Provider      string `json:"provider"`
}

func (q *Queries) CreatePaymentMethod(ctx context.Context, arg CreatePaymentMethodParams) (PaymentMethod, error) {
	row := q.db.QueryRow(ctx, createPaymentMethod, arg.UserID, arg.PaymentTypeID, arg.Provider)
	var i PaymentMethod
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentTypeID,
		&i.Provider,
		&i.IsDefault,
	)
	return i, err
}

const deletePaymentMethod = `-- name: DeletePaymentMethod :one
DELETE FROM "payment_method"
WHERE id = $1
AND user_id = $2
RETURNING id, user_id, payment_type_id, provider, is_default
`

type DeletePaymentMethodParams struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) DeletePaymentMethod(ctx context.Context, arg DeletePaymentMethodParams) (PaymentMethod, error) {
	row := q.db.QueryRow(ctx, deletePaymentMethod, arg.ID, arg.UserID)
	var i PaymentMethod
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentTypeID,
		&i.Provider,
		&i.IsDefault,
	)
	return i, err
}

const getPaymentMethod = `-- name: GetPaymentMethod :one
SELECT id, user_id, payment_type_id, provider, is_default FROM "payment_method"
WHERE id = $1 
AND user_id = $2
LIMIT 1
`

type GetPaymentMethodParams struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) GetPaymentMethod(ctx context.Context, arg GetPaymentMethodParams) (PaymentMethod, error) {
	row := q.db.QueryRow(ctx, getPaymentMethod, arg.ID, arg.UserID)
	var i PaymentMethod
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentTypeID,
		&i.Provider,
		&i.IsDefault,
	)
	return i, err
}

const listPaymentMethods = `-- name: ListPaymentMethods :many
SELECT id, user_id, payment_type_id, provider, is_default FROM "payment_method"
WHERE user_id = $3
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListPaymentMethodsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) ListPaymentMethods(ctx context.Context, arg ListPaymentMethodsParams) ([]PaymentMethod, error) {
	rows, err := q.db.Query(ctx, listPaymentMethods, arg.Limit, arg.Offset, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PaymentMethod{}
	for rows.Next() {
		var i PaymentMethod
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.PaymentTypeID,
			&i.Provider,
			&i.IsDefault,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePaymentMethod = `-- name: UpdatePaymentMethod :one
UPDATE "payment_method"
SET 
user_id = COALESCE($1,user_id),
payment_type_id = COALESCE($2,payment_type_id),
provider = COALESCE($3,provider)
WHERE id = $4
RETURNING id, user_id, payment_type_id, provider, is_default
`

type UpdatePaymentMethodParams struct {
	UserID        null.Int    `json:"user_id"`
	PaymentTypeID null.Int    `json:"payment_type_id"`
	Provider      null.String `json:"provider"`
	ID            int64       `json:"id"`
}

func (q *Queries) UpdatePaymentMethod(ctx context.Context, arg UpdatePaymentMethodParams) (PaymentMethod, error) {
	row := q.db.QueryRow(ctx, updatePaymentMethod,
		arg.UserID,
		arg.PaymentTypeID,
		arg.Provider,
		arg.ID,
	)
	var i PaymentMethod
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentTypeID,
		&i.Provider,
		&i.IsDefault,
	)
	return i, err
}
