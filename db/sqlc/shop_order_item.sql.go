// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: shop_order_item.sql

package db

import (
	"context"
	"database/sql"
)

const createShopOrderItem = `-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  quantity,
  price
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, product_item_id, order_id, quantity, price, created_at, updated_at
`

type CreateShopOrderItemParams struct {
	ProductItemID int64  `json:"product_item_id"`
	OrderID       int64  `json:"order_id"`
	Quantity      int32  `json:"quantity"`
	Price         string `json:"price"`
}

func (q *Queries) CreateShopOrderItem(ctx context.Context, arg CreateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRowContext(ctx, createShopOrderItem,
		arg.ProductItemID,
		arg.OrderID,
		arg.Quantity,
		arg.Price,
	)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteShopOrderItem = `-- name: DeleteShopOrderItem :exec
DELETE FROM "shop_order_item"
WHERE id = $1
`

func (q *Queries) DeleteShopOrderItem(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteShopOrderItem, id)
	return err
}

const getShopOrderItem = `-- name: GetShopOrderItem :one
SELECT id, product_item_id, order_id, quantity, price, created_at, updated_at FROM "shop_order_item"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error) {
	row := q.db.QueryRowContext(ctx, getShopOrderItem, id)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listShopOrderItems = `-- name: ListShopOrderItems :many
SELECT id, product_item_id, order_id, quantity, price, created_at, updated_at FROM "shop_order_item"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListShopOrderItemsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListShopOrderItems(ctx context.Context, arg ListShopOrderItemsParams) ([]ShopOrderItem, error) {
	rows, err := q.db.QueryContext(ctx, listShopOrderItems, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShopOrderItem{}
	for rows.Next() {
		var i ShopOrderItem
		if err := rows.Scan(
			&i.ID,
			&i.ProductItemID,
			&i.OrderID,
			&i.Quantity,
			&i.Price,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listShopOrderItemsByOrderID = `-- name: ListShopOrderItemsByOrderID :many
SELECT id, product_item_id, order_id, quantity, price, created_at, updated_at FROM "shop_order_item"
WHERE order_id = $1
ORDER BY id
`

func (q *Queries) ListShopOrderItemsByOrderID(ctx context.Context, orderID int64) ([]ShopOrderItem, error) {
	rows, err := q.db.QueryContext(ctx, listShopOrderItemsByOrderID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShopOrderItem{}
	for rows.Next() {
		var i ShopOrderItem
		if err := rows.Scan(
			&i.ID,
			&i.ProductItemID,
			&i.OrderID,
			&i.Quantity,
			&i.Price,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateShopOrderItem = `-- name: UpdateShopOrderItem :one
UPDATE "shop_order_item"
SET 
product_item_id = COALESCE($1,product_item_id),
order_id = COALESCE($2,order_id),
quantity = COALESCE($3,quantity),
price = COALESCE($4,price)
WHERE id = $5
RETURNING id, product_item_id, order_id, quantity, price, created_at, updated_at
`

type UpdateShopOrderItemParams struct {
	ProductItemID sql.NullInt64  `json:"product_item_id"`
	OrderID       sql.NullInt64  `json:"order_id"`
	Quantity      sql.NullInt32  `json:"quantity"`
	Price         sql.NullString `json:"price"`
	ID            int64          `json:"id"`
}

func (q *Queries) UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRowContext(ctx, updateShopOrderItem,
		arg.ProductItemID,
		arg.OrderID,
		arg.Quantity,
		arg.Price,
		arg.ID,
	)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
