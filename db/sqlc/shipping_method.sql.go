// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: shipping_method.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createShippingMethod = `-- name: CreateShippingMethod :one
INSERT INTO "shipping_method" (
  name,
  price
) VALUES (
  $1, $2
)
RETURNING id, name, price
`

type CreateShippingMethodParams struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}

func (q *Queries) CreateShippingMethod(ctx context.Context, arg CreateShippingMethodParams) (ShippingMethod, error) {
	row := q.db.QueryRow(ctx, createShippingMethod, arg.Name, arg.Price)
	var i ShippingMethod
	err := row.Scan(&i.ID, &i.Name, &i.Price)
	return i, err
}

const deleteShippingMethod = `-- name: DeleteShippingMethod :exec
DELETE FROM "shipping_method"
WHERE id = $1
`

func (q *Queries) DeleteShippingMethod(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteShippingMethod, id)
	return err
}

const getShippingMethod = `-- name: GetShippingMethod :one
SELECT id, name, price FROM "shipping_method"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShippingMethod(ctx context.Context, id int64) (ShippingMethod, error) {
	row := q.db.QueryRow(ctx, getShippingMethod, id)
	var i ShippingMethod
	err := row.Scan(&i.ID, &i.Name, &i.Price)
	return i, err
}

const listShippingMethods = `-- name: ListShippingMethods :many
SELECT id, name, price FROM "shipping_method"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListShippingMethodsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListShippingMethods(ctx context.Context, arg ListShippingMethodsParams) ([]ShippingMethod, error) {
	rows, err := q.db.Query(ctx, listShippingMethods, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShippingMethod{}
	for rows.Next() {
		var i ShippingMethod
		if err := rows.Scan(&i.ID, &i.Name, &i.Price); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateShippingMethod = `-- name: UpdateShippingMethod :one
UPDATE "shipping_method"
SET 
name = COALESCE($1,name),
price = COALESCE($2,price)
WHERE id = $3
RETURNING id, name, price
`

type UpdateShippingMethodParams struct {
	Name  null.String `json:"name"`
	Price null.String `json:"price"`
	ID    int64       `json:"id"`
}

func (q *Queries) UpdateShippingMethod(ctx context.Context, arg UpdateShippingMethodParams) (ShippingMethod, error) {
	row := q.db.QueryRow(ctx, updateShippingMethod, arg.Name, arg.Price, arg.ID)
	var i ShippingMethod
	err := row.Scan(&i.ID, &i.Name, &i.Price)
	return i, err
}
