// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: payment_type.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createPaymentType = `-- name: CreatePaymentType :one
INSERT INTO "payment_type" (
  value
) VALUES (
  $1
) 
ON CONFLICT(value) DO UPDATE SET value = $1
RETURNING id, value, is_active
`

func (q *Queries) CreatePaymentType(ctx context.Context, value string) (PaymentType, error) {
	row := q.db.QueryRow(ctx, createPaymentType, value)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value, &i.IsActive)
	return i, err
}

const deletePaymentType = `-- name: DeletePaymentType :exec
DELETE FROM "payment_type"
WHERE id = $1
`

func (q *Queries) DeletePaymentType(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deletePaymentType, id)
	return err
}

const getPaymentType = `-- name: GetPaymentType :one
SELECT id, value, is_active FROM "payment_type"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetPaymentType(ctx context.Context, id int64) (PaymentType, error) {
	row := q.db.QueryRow(ctx, getPaymentType, id)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value, &i.IsActive)
	return i, err
}

const listPaymentTypes = `-- name: ListPaymentTypes :many
SELECT id, value, is_active FROM "payment_type"
`

func (q *Queries) ListPaymentTypes(ctx context.Context) ([]PaymentType, error) {
	rows, err := q.db.Query(ctx, listPaymentTypes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PaymentType{}
	for rows.Next() {
		var i PaymentType
		if err := rows.Scan(&i.ID, &i.Value, &i.IsActive); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePaymentType = `-- name: UpdatePaymentType :one

UPDATE "payment_type"
SET 
value = COALESCE($1,value),
is_active = COALESCE($2,is_active)
WHERE id = $3
RETURNING id, value, is_active
`

type UpdatePaymentTypeParams struct {
	Value    null.String `json:"value"`
	IsActive null.Bool   `json:"is_active"`
	ID       int64       `json:"id"`
}

// ORDER BY id
// LIMIT $1
// OFFSET $2;
func (q *Queries) UpdatePaymentType(ctx context.Context, arg UpdatePaymentTypeParams) (PaymentType, error) {
	row := q.db.QueryRow(ctx, updatePaymentType, arg.Value, arg.IsActive, arg.ID)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value, &i.IsActive)
	return i, err
}
