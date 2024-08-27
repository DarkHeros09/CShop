// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: shopping_cart _item.sql

package db

import (
	"context"
	"time"

	null "github.com/guregu/null/v5"
)

const createShoppingCartItem = `-- name: CreateShoppingCartItem :one
INSERT INTO "shopping_cart_item" (
  shopping_cart_id,
  product_item_id,
  qty
) VALUES (
  $1, $2, $3
)
RETURNING id, shopping_cart_id, product_item_id, qty, created_at, updated_at
`

type CreateShoppingCartItemParams struct {
	ShoppingCartID int64 `json:"shopping_cart_id"`
	ProductItemID  int64 `json:"product_item_id"`
	Qty            int32 `json:"qty"`
}

func (q *Queries) CreateShoppingCartItem(ctx context.Context, arg CreateShoppingCartItemParams) (ShoppingCartItem, error) {
	row := q.db.QueryRow(ctx, createShoppingCartItem, arg.ShoppingCartID, arg.ProductItemID, arg.Qty)
	var i ShoppingCartItem
	err := row.Scan(
		&i.ID,
		&i.ShoppingCartID,
		&i.ProductItemID,
		&i.Qty,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteShoppingCartItem = `-- name: DeleteShoppingCartItem :exec
WITH t1 AS (
  SELECT id FROM "shopping_cart" AS sc
  WHERE sc.user_id = $2
  AND sc.id = $3
)
DELETE FROM "shopping_cart_item" AS sci
WHERE sci.id = $1
AND sci.shopping_cart_id = (SELECT id FROM t1)
`

type DeleteShoppingCartItemParams struct {
	ShoppingCartItemID int64 `json:"shopping_cart_item_id"`
	UserID             int64 `json:"user_id"`
	ShoppingCartID     int64 `json:"shopping_cart_id"`
}

func (q *Queries) DeleteShoppingCartItem(ctx context.Context, arg DeleteShoppingCartItemParams) error {
	_, err := q.db.Exec(ctx, deleteShoppingCartItem, arg.ShoppingCartItemID, arg.UserID, arg.ShoppingCartID)
	return err
}

const deleteShoppingCartItemAllByUser = `-- name: DeleteShoppingCartItemAllByUser :many
WITH t1 AS(
  SELECT id FROM "shopping_cart" AS sc 
  WHERE sc.user_id = $1
  AND sc.id = $2
)
DELETE FROM "shopping_cart_item"
WHERE shopping_cart_id = (SELECT id FROM t1)
RETURNING id, shopping_cart_id, product_item_id, qty, created_at, updated_at
`

type DeleteShoppingCartItemAllByUserParams struct {
	UserID         int64 `json:"user_id"`
	ShoppingCartID int64 `json:"shopping_cart_id"`
}

func (q *Queries) DeleteShoppingCartItemAllByUser(ctx context.Context, arg DeleteShoppingCartItemAllByUserParams) ([]ShoppingCartItem, error) {
	rows, err := q.db.Query(ctx, deleteShoppingCartItemAllByUser, arg.UserID, arg.ShoppingCartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShoppingCartItem{}
	for rows.Next() {
		var i ShoppingCartItem
		if err := rows.Scan(
			&i.ID,
			&i.ShoppingCartID,
			&i.ProductItemID,
			&i.Qty,
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

const getShoppingCartItem = `-- name: GetShoppingCartItem :one
SELECT id, shopping_cart_id, product_item_id, qty, created_at, updated_at FROM "shopping_cart_item"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShoppingCartItem(ctx context.Context, id int64) (ShoppingCartItem, error) {
	row := q.db.QueryRow(ctx, getShoppingCartItem, id)
	var i ShoppingCartItem
	err := row.Scan(
		&i.ID,
		&i.ShoppingCartID,
		&i.ProductItemID,
		&i.Qty,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getShoppingCartItemByUserIDCartID = `-- name: GetShoppingCartItemByUserIDCartID :many
SELECT sci.id, sci.shopping_cart_id, sci.product_item_id, sci.qty, sci.created_at, sci.updated_at, sc.user_id
FROM "shopping_cart_item" AS sci
LEFT JOIN "shopping_cart" AS sc ON sc.id = sci.shopping_cart_id
WHERE sc.user_id = $1
AND sc.id = $2
`

type GetShoppingCartItemByUserIDCartIDParams struct {
	UserID int64 `json:"user_id"`
	ID     int64 `json:"id"`
}

type GetShoppingCartItemByUserIDCartIDRow struct {
	ID             int64     `json:"id"`
	ShoppingCartID int64     `json:"shopping_cart_id"`
	ProductItemID  int64     `json:"product_item_id"`
	Qty            int32     `json:"qty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserID         null.Int  `json:"user_id"`
}

func (q *Queries) GetShoppingCartItemByUserIDCartID(ctx context.Context, arg GetShoppingCartItemByUserIDCartIDParams) ([]GetShoppingCartItemByUserIDCartIDRow, error) {
	rows, err := q.db.Query(ctx, getShoppingCartItemByUserIDCartID, arg.UserID, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetShoppingCartItemByUserIDCartIDRow{}
	for rows.Next() {
		var i GetShoppingCartItemByUserIDCartIDRow
		if err := rows.Scan(
			&i.ID,
			&i.ShoppingCartID,
			&i.ProductItemID,
			&i.Qty,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const listShoppingCartItems = `-- name: ListShoppingCartItems :many

SELECT sci.id, sci.shopping_cart_id, sci.product_item_id, sci.qty, sci.created_at, sci.updated_at FROM "shopping_cart_item" AS sci
ORDER BY sci.id
LIMIT $1
OFFSET $2
`

type ListShoppingCartItemsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// LIMIT 1;
func (q *Queries) ListShoppingCartItems(ctx context.Context, arg ListShoppingCartItemsParams) ([]ShoppingCartItem, error) {
	rows, err := q.db.Query(ctx, listShoppingCartItems, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShoppingCartItem{}
	for rows.Next() {
		var i ShoppingCartItem
		if err := rows.Scan(
			&i.ID,
			&i.ShoppingCartID,
			&i.ProductItemID,
			&i.Qty,
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

const listShoppingCartItemsByCartID = `-- name: ListShoppingCartItemsByCartID :many
SELECT id, shopping_cart_id, product_item_id, qty, created_at, updated_at FROM "shopping_cart_item"
WHERE shopping_cart_id = $1
ORDER BY id
`

func (q *Queries) ListShoppingCartItemsByCartID(ctx context.Context, shoppingCartID int64) ([]ShoppingCartItem, error) {
	rows, err := q.db.Query(ctx, listShoppingCartItemsByCartID, shoppingCartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShoppingCartItem{}
	for rows.Next() {
		var i ShoppingCartItem
		if err := rows.Scan(
			&i.ID,
			&i.ShoppingCartID,
			&i.ProductItemID,
			&i.Qty,
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

const listShoppingCartItemsByUserID = `-- name: ListShoppingCartItemsByUserID :many
SELECT sc.user_id, sci.id, sci.shopping_cart_id, sci.product_item_id, sci.qty, sci.created_at, sci.updated_at
FROM "shopping_cart" AS sc
LEFT JOIN "shopping_cart_item" AS sci ON sci.shopping_cart_id = sc.id
WHERE sc.user_id = $1
`

type ListShoppingCartItemsByUserIDRow struct {
	UserID         int64     `json:"user_id"`
	ID             null.Int  `json:"id"`
	ShoppingCartID null.Int  `json:"shopping_cart_id"`
	ProductItemID  null.Int  `json:"product_item_id"`
	Qty            null.Int  `json:"qty"`
	CreatedAt      null.Time `json:"created_at"`
	UpdatedAt      null.Time `json:"updated_at"`
}

func (q *Queries) ListShoppingCartItemsByUserID(ctx context.Context, userID int64) ([]ListShoppingCartItemsByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listShoppingCartItemsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListShoppingCartItemsByUserIDRow{}
	for rows.Next() {
		var i ListShoppingCartItemsByUserIDRow
		if err := rows.Scan(
			&i.UserID,
			&i.ID,
			&i.ShoppingCartID,
			&i.ProductItemID,
			&i.Qty,
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

const updateShoppingCartItem = `-- name: UpdateShoppingCartItem :one
WITH t1 AS (
  SELECT user_id FROM "shopping_cart" AS sc
  WHERE sc.id = $4
)

UPDATE "shopping_cart_item" AS sci
SET 
product_item_id = COALESCE($1,product_item_id),
qty = COALESCE($2,qty),
updated_at = now()
WHERE sci.id = $3
RETURNING id, shopping_cart_id, product_item_id, qty, created_at, updated_at, (SELECT user_id FROM t1)
`

type UpdateShoppingCartItemParams struct {
	ProductItemID  null.Int `json:"product_item_id"`
	Qty            null.Int `json:"qty"`
	ID             int64    `json:"id"`
	ShoppingCartID int64    `json:"shopping_cart_id"`
}

type UpdateShoppingCartItemRow struct {
	ID             int64     `json:"id"`
	ShoppingCartID int64     `json:"shopping_cart_id"`
	ProductItemID  int64     `json:"product_item_id"`
	Qty            int32     `json:"qty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserID         int64     `json:"user_id"`
}

func (q *Queries) UpdateShoppingCartItem(ctx context.Context, arg UpdateShoppingCartItemParams) (UpdateShoppingCartItemRow, error) {
	row := q.db.QueryRow(ctx, updateShoppingCartItem,
		arg.ProductItemID,
		arg.Qty,
		arg.ID,
		arg.ShoppingCartID,
	)
	var i UpdateShoppingCartItemRow
	err := row.Scan(
		&i.ID,
		&i.ShoppingCartID,
		&i.ProductItemID,
		&i.Qty,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}
