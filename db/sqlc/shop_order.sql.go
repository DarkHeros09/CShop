// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: shop_order.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createShopOrder = `-- name: CreateShopOrder :one
INSERT INTO "shop_order" (
  user_id,
  payment_method_id,
  shipping_address_id,
  order_total,
  shipping_method_id,
  order_status_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at
`

type CreateShopOrderParams struct {
	UserID            int64  `json:"user_id"`
	PaymentMethodID   int64  `json:"payment_method_id"`
	ShippingAddressID int64  `json:"shipping_address_id"`
	OrderTotal        string `json:"order_total"`
	ShippingMethodID  int64  `json:"shipping_method_id"`
	OrderStatusID     int64  `json:"order_status_id"`
}

func (q *Queries) CreateShopOrder(ctx context.Context, arg CreateShopOrderParams) (ShopOrder, error) {
	row := q.db.QueryRow(ctx, createShopOrder,
		arg.UserID,
		arg.PaymentMethodID,
		arg.ShippingAddressID,
		arg.OrderTotal,
		arg.ShippingMethodID,
		arg.OrderStatusID,
	)
	var i ShopOrder
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentMethodID,
		&i.ShippingAddressID,
		&i.OrderTotal,
		&i.ShippingMethodID,
		&i.OrderStatusID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteShopOrder = `-- name: DeleteShopOrder :exec
DELETE FROM "shop_order"
WHERE id = $1
`

func (q *Queries) DeleteShopOrder(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteShopOrder, id)
	return err
}

const getShopOrder = `-- name: GetShopOrder :one
SELECT id, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at FROM "shop_order"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShopOrder(ctx context.Context, id int64) (ShopOrder, error) {
	row := q.db.QueryRow(ctx, getShopOrder, id)
	var i ShopOrder
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentMethodID,
		&i.ShippingAddressID,
		&i.OrderTotal,
		&i.ShippingMethodID,
		&i.OrderStatusID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listShopOrders = `-- name: ListShopOrders :many
SELECT id, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at FROM "shop_order"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListShopOrdersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListShopOrders(ctx context.Context, arg ListShopOrdersParams) ([]ShopOrder, error) {
	rows, err := q.db.Query(ctx, listShopOrders, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ShopOrder{}
	for rows.Next() {
		var i ShopOrder
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.PaymentMethodID,
			&i.ShippingAddressID,
			&i.OrderTotal,
			&i.ShippingMethodID,
			&i.OrderStatusID,
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

const updateShopOrder = `-- name: UpdateShopOrder :one
UPDATE "shop_order"
SET 
user_id = COALESCE($1,user_id),
payment_method_id = COALESCE($2,payment_method_id),
shipping_address_id = COALESCE($3,shipping_address_id),
order_total = COALESCE($4,order_total),
shipping_method_id = COALESCE($5,shipping_method_id),
order_status_id = COALESCE($6,order_status_id)
WHERE id = $7
RETURNING id, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at
`

type UpdateShopOrderParams struct {
	UserID            null.Int    `json:"user_id"`
	PaymentMethodID   null.Int    `json:"payment_method_id"`
	ShippingAddressID null.Int    `json:"shipping_address_id"`
	OrderTotal        null.String `json:"order_total"`
	ShippingMethodID  null.Int    `json:"shipping_method_id"`
	OrderStatusID     null.Int    `json:"order_status_id"`
	ID                int64       `json:"id"`
}

func (q *Queries) UpdateShopOrder(ctx context.Context, arg UpdateShopOrderParams) (ShopOrder, error) {
	row := q.db.QueryRow(ctx, updateShopOrder,
		arg.UserID,
		arg.PaymentMethodID,
		arg.ShippingAddressID,
		arg.OrderTotal,
		arg.ShippingMethodID,
		arg.OrderStatusID,
		arg.ID,
	)
	var i ShopOrder
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PaymentMethodID,
		&i.ShippingAddressID,
		&i.OrderTotal,
		&i.ShippingMethodID,
		&i.OrderStatusID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
