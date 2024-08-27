// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: shipping_method.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const createShippingMethod = `-- name: CreateShippingMethod :one
INSERT INTO "shipping_method" (
  name,
  price
) VALUES (
  $1, $2
)
ON CONFLICT (name) DO UPDATE SET 
name = EXCLUDED.name,
price = EXCLUDED.price
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

const getShippingMethodByUserID = `-- name: GetShippingMethodByUserID :one
SELECT sm.id, sm.name, sm.price, so.user_id
FROM "shipping_method" AS sm
LEFT JOIN "shop_order" AS so ON so.shipping_method_id = sm.id
WHERE so.user_id = $1
AND sm.id = $2
LIMIT 1
`

type GetShippingMethodByUserIDParams struct {
	UserID int64 `json:"user_id"`
	ID     int64 `json:"id"`
}

type GetShippingMethodByUserIDRow struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Price  string   `json:"price"`
	UserID null.Int `json:"user_id"`
}

func (q *Queries) GetShippingMethodByUserID(ctx context.Context, arg GetShippingMethodByUserIDParams) (GetShippingMethodByUserIDRow, error) {
	row := q.db.QueryRow(ctx, getShippingMethodByUserID, arg.UserID, arg.ID)
	var i GetShippingMethodByUserIDRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.UserID,
	)
	return i, err
}

const listShippingMethods = `-- name: ListShippingMethods :many
SELECT id, name, price FROM "shipping_method"
`

func (q *Queries) ListShippingMethods(ctx context.Context) ([]ShippingMethod, error) {
	rows, err := q.db.Query(ctx, listShippingMethods)
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

const listShippingMethodsByUserID = `-- name: ListShippingMethodsByUserID :many

SELECT sm.id, sm.name, sm.price, so.user_id
FROM "shipping_method" AS sm
LEFT JOIN "shop_order" AS so ON so.shipping_method_id = sm.id
WHERE so.user_id = $3
ORDER BY sm.id
LIMIT $1
OFFSET $2
`

type ListShippingMethodsByUserIDParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserID int64 `json:"user_id"`
}

type ListShippingMethodsByUserIDRow struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Price  string   `json:"price"`
	UserID null.Int `json:"user_id"`
}

// ORDER BY id
// LIMIT $1
// OFFSET $2;
func (q *Queries) ListShippingMethodsByUserID(ctx context.Context, arg ListShippingMethodsByUserIDParams) ([]ListShippingMethodsByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listShippingMethodsByUserID, arg.Limit, arg.Offset, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListShippingMethodsByUserIDRow{}
	for rows.Next() {
		var i ListShippingMethodsByUserIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.UserID,
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
