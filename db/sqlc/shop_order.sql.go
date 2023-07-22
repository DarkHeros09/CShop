// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: shop_order.sql

package db

import (
	"context"
	"time"

	"github.com/guregu/null"
)

const createShopOrder = `-- name: CreateShopOrder :one
INSERT INTO "shop_order" (
  track_number,
  user_id,
  payment_method_id,
  shipping_address_id,
  order_total,
  shipping_method_id,
  order_status_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, track_number, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at
`

type CreateShopOrderParams struct {
	TrackNumber       string   `json:"track_number"`
	UserID            int64    `json:"user_id"`
	PaymentMethodID   int64    `json:"payment_method_id"`
	ShippingAddressID int64    `json:"shipping_address_id"`
	OrderTotal        string   `json:"order_total"`
	ShippingMethodID  int64    `json:"shipping_method_id"`
	OrderStatusID     null.Int `json:"order_status_id"`
}

func (q *Queries) CreateShopOrder(ctx context.Context, arg CreateShopOrderParams) (ShopOrder, error) {
	row := q.db.QueryRow(ctx, createShopOrder,
		arg.TrackNumber,
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
		&i.TrackNumber,
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
SELECT id, track_number, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at FROM "shop_order"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShopOrder(ctx context.Context, id int64) (ShopOrder, error) {
	row := q.db.QueryRow(ctx, getShopOrder, id)
	var i ShopOrder
	err := row.Scan(
		&i.ID,
		&i.TrackNumber,
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
SELECT id, track_number, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at FROM "shop_order"
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
			&i.TrackNumber,
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

const listShopOrdersByUserID = `-- name: ListShopOrdersByUserID :many
SELECT os.status,
ROW_NUMBER() OVER(ORDER BY so.id) as order_number,
(
  SELECT COUNT(soi.id) FROM "shop_order_item" AS soi
  WHERE soi.order_id = so.id
) AS item_count,so.id, so.track_number, so.user_id, so.payment_method_id, so.shipping_address_id, so.order_total, so.shipping_method_id, so.order_status_id, so.created_at, so.updated_at
FROM "shop_order" AS so
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
WHERE so.user_id = $1
ORDER BY so.id DESC
LIMIT $2
OFFSET $3
`

type ListShopOrdersByUserIDParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListShopOrdersByUserIDRow struct {
	Status            null.String `json:"status"`
	OrderNumber       int64       `json:"order_number"`
	ItemCount         int64       `json:"item_count"`
	ID                int64       `json:"id"`
	TrackNumber       string      `json:"track_number"`
	UserID            int64       `json:"user_id"`
	PaymentMethodID   int64       `json:"payment_method_id"`
	ShippingAddressID int64       `json:"shipping_address_id"`
	OrderTotal        string      `json:"order_total"`
	ShippingMethodID  int64       `json:"shipping_method_id"`
	OrderStatusID     null.Int    `json:"order_status_id"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

func (q *Queries) ListShopOrdersByUserID(ctx context.Context, arg ListShopOrdersByUserIDParams) ([]ListShopOrdersByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listShopOrdersByUserID, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListShopOrdersByUserIDRow{}
	for rows.Next() {
		var i ListShopOrdersByUserIDRow
		if err := rows.Scan(
			&i.Status,
			&i.OrderNumber,
			&i.ItemCount,
			&i.ID,
			&i.TrackNumber,
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
track_number = COALESCE($1,track_number),
user_id = COALESCE($2,user_id),
payment_method_id = COALESCE($3,payment_method_id),
shipping_address_id = COALESCE($4,shipping_address_id),
order_total = COALESCE($5,order_total),
shipping_method_id = COALESCE($6,shipping_method_id),
order_status_id = COALESCE($7,order_status_id),
updated_at = now()
WHERE id = $8
RETURNING id, track_number, user_id, payment_method_id, shipping_address_id, order_total, shipping_method_id, order_status_id, created_at, updated_at
`

type UpdateShopOrderParams struct {
	TrackNumber       null.String `json:"track_number"`
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
		arg.TrackNumber,
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
		&i.TrackNumber,
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
