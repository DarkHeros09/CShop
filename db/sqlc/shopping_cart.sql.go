// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: shopping_cart.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createShoppingCart = `-- name: CreateShoppingCart :one
INSERT INTO "shopping_cart" (
  user_id
) VALUES (
  $1
)
RETURNING id, user_id, created_at, updated_at
`

func (q *Queries) CreateShoppingCart(ctx context.Context, userID int64) (ShoppingCart, error) {
	row := q.db.QueryRow(ctx, createShoppingCart, userID)
	var i ShoppingCart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteShoppingCart = `-- name: DeleteShoppingCart :exec
DELETE FROM "shopping_cart"
WHERE id = $1
`

func (q *Queries) DeleteShoppingCart(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteShoppingCart, id)
	return err
}

const getShoppingCart = `-- name: GetShoppingCart :one
SELECT id, user_id, created_at, updated_at FROM "shopping_cart"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShoppingCart(ctx context.Context, id int64) (ShoppingCart, error) {
	row := q.db.QueryRow(ctx, getShoppingCart, id)
	var i ShoppingCart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getShoppingCartByUserIDCartID = `-- name: GetShoppingCartByUserIDCartID :one
SELECT id, user_id, created_at, updated_at FROM "shopping_cart"
WHERE user_id = $1
AND id = $2
LIMIT 1
`

type GetShoppingCartByUserIDCartIDParams struct {
	UserID int64 `json:"user_id"`
	ID     int64 `json:"id"`
}

func (q *Queries) GetShoppingCartByUserIDCartID(ctx context.Context, arg GetShoppingCartByUserIDCartIDParams) (ShoppingCart, error) {
	row := q.db.QueryRow(ctx, getShoppingCartByUserIDCartID, arg.UserID, arg.ID)
	var i ShoppingCart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listShoppingCarts = `-- name: ListShoppingCarts :many
SELECT id, user_id, created_at, updated_at FROM "shopping_cart"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListShoppingCartsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListShoppingCarts(ctx context.Context, arg ListShoppingCartsParams) ([]ShoppingCart, error) {
	rows, err := q.db.Query(ctx, listShoppingCarts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShoppingCart{}
	for rows.Next() {
		var i ShoppingCart
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const updateShoppingCart = `-- name: UpdateShoppingCart :one
UPDATE "shopping_cart"
SET 
user_id = COALESCE($1,user_id),
updated_at = now()
WHERE id = $2
RETURNING id, user_id, created_at, updated_at
`

type UpdateShoppingCartParams struct {
	UserID null.Int `json:"user_id"`
	ID     int64    `json:"id"`
}

func (q *Queries) UpdateShoppingCart(ctx context.Context, arg UpdateShoppingCartParams) (ShoppingCart, error) {
	row := q.db.QueryRow(ctx, updateShoppingCart, arg.UserID, arg.ID)
	var i ShoppingCart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
