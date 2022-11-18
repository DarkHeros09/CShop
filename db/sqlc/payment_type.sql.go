// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
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
RETURNING id, value
`

func (q *Queries) CreatePaymentType(ctx context.Context, value string) (PaymentType, error) {
	row := q.db.QueryRow(ctx, createPaymentType, value)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value)
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
SELECT id, value FROM "payment_type"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetPaymentType(ctx context.Context, id int64) (PaymentType, error) {
	row := q.db.QueryRow(ctx, getPaymentType, id)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value)
	return i, err
}

const listPaymentTypes = `-- name: ListPaymentTypes :many
SELECT id, value FROM "payment_type"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListPaymentTypesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListPaymentTypes(ctx context.Context, arg ListPaymentTypesParams) ([]PaymentType, error) {
	rows, err := q.db.Query(ctx, listPaymentTypes, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PaymentType{}
	for rows.Next() {
		var i PaymentType
		if err := rows.Scan(&i.ID, &i.Value); err != nil {
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
value = COALESCE($1,value)
WHERE id = $2
RETURNING id, value
`

type UpdatePaymentTypeParams struct {
	Value null.String `json:"value"`
	ID    int64       `json:"id"`
}

func (q *Queries) UpdatePaymentType(ctx context.Context, arg UpdatePaymentTypeParams) (PaymentType, error) {
	row := q.db.QueryRow(ctx, updatePaymentType, arg.Value, arg.ID)
	var i PaymentType
	err := row.Scan(&i.ID, &i.Value)
	return i, err
}
