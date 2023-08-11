// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: shop_order_item.sql

package db

import (
	"context"
	"time"

	"github.com/guregu/null"
)

const createShopOrderItem = `-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  size,
  color,
  quantity,
  price
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, product_item_id, order_id, size, color, quantity, price, created_at, updated_at
`

type CreateShopOrderItemParams struct {
	ProductItemID int64  `json:"product_item_id"`
	OrderID       int64  `json:"order_id"`
	Size          string `json:"size"`
	Color         string `json:"color"`
	Quantity      int32  `json:"quantity"`
	Price         string `json:"price"`
}

func (q *Queries) CreateShopOrderItem(ctx context.Context, arg CreateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, createShopOrderItem,
		arg.ProductItemID,
		arg.OrderID,
		arg.Size,
		arg.Color,
		arg.Quantity,
		arg.Price,
	)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Size,
		&i.Color,
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
	_, err := q.db.Exec(ctx, deleteShopOrderItem, id)
	return err
}

const getShopOrderItem = `-- name: GetShopOrderItem :one
SELECT id, product_item_id, order_id, size, color, quantity, price, created_at, updated_at FROM "shop_order_item"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, getShopOrderItem, id)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Size,
		&i.Color,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getShopOrderItemByUserIDOrderID = `-- name: GetShopOrderItemByUserIDOrderID :one
SELECT soi.id, soi.product_item_id, soi.order_id, soi.size, soi.color, soi.quantity, soi.price, soi.created_at, soi.updated_at, so.user_id
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
WHERE so.user_id = $1
AND soi.order_id = $2 
LIMIT 1
`

type GetShopOrderItemByUserIDOrderIDParams struct {
	UserID  int64 `json:"user_id"`
	OrderID int64 `json:"order_id"`
}

type GetShopOrderItemByUserIDOrderIDRow struct {
	ID            int64     `json:"id"`
	ProductItemID int64     `json:"product_item_id"`
	OrderID       int64     `json:"order_id"`
	Size          string    `json:"size"`
	Color         string    `json:"color"`
	Quantity      int32     `json:"quantity"`
	Price         string    `json:"price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	UserID        null.Int  `json:"user_id"`
}

func (q *Queries) GetShopOrderItemByUserIDOrderID(ctx context.Context, arg GetShopOrderItemByUserIDOrderIDParams) (GetShopOrderItemByUserIDOrderIDRow, error) {
	row := q.db.QueryRow(ctx, getShopOrderItemByUserIDOrderID, arg.UserID, arg.OrderID)
	var i GetShopOrderItemByUserIDOrderIDRow
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Size,
		&i.Color,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}

const listShopOrderItems = `-- name: ListShopOrderItems :many
SELECT id, product_item_id, order_id, size, color, quantity, price, created_at, updated_at FROM "shop_order_item"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListShopOrderItemsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListShopOrderItems(ctx context.Context, arg ListShopOrderItemsParams) ([]ShopOrderItem, error) {
	rows, err := q.db.Query(ctx, listShopOrderItems, arg.Limit, arg.Offset)
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
			&i.Size,
			&i.Color,
			&i.Quantity,
			&i.Price,
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

const listShopOrderItemsByUserID = `-- name: ListShopOrderItemsByUserID :many

SELECT so.id, so.track_number, so.user_id, so.payment_method_id, so.shipping_address_id, so.order_total, so.shipping_method_id, so.order_status_id, so.created_at, so.updated_at, soi.id, soi.product_item_id, soi.order_id, soi.size, soi.color, soi.quantity, soi.price, soi.created_at, soi.updated_at 
FROM "shop_order" AS so
LEFT JOIN "shop_order_item" AS soi ON soi.order_id = so.id
WHERE so.user_id = $3
ORDER BY so.id
LIMIT $1
OFFSET $2
`

type ListShopOrderItemsByUserIDParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserID int64 `json:"user_id"`
}

type ListShopOrderItemsByUserIDRow struct {
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
	ID_2              null.Int    `json:"id_2"`
	ProductItemID     null.Int    `json:"product_item_id"`
	OrderID           null.Int    `json:"order_id"`
	Size              null.String `json:"size"`
	Color             null.String `json:"color"`
	Quantity          null.Int    `json:"quantity"`
	Price             null.String `json:"price"`
	CreatedAt_2       null.Time   `json:"created_at_2"`
	UpdatedAt_2       null.Time   `json:"updated_at_2"`
}

// ORDER BY soi.id;
func (q *Queries) ListShopOrderItemsByUserID(ctx context.Context, arg ListShopOrderItemsByUserIDParams) ([]ListShopOrderItemsByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listShopOrderItemsByUserID, arg.Limit, arg.Offset, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListShopOrderItemsByUserIDRow{}
	for rows.Next() {
		var i ListShopOrderItemsByUserIDRow
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
			&i.ID_2,
			&i.ProductItemID,
			&i.OrderID,
			&i.Size,
			&i.Color,
			&i.Quantity,
			&i.Price,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
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

const listShopOrderItemsByUserIDOrderID = `-- name: ListShopOrderItemsByUserIDOrderID :many
SELECT os.status, so.track_number, sm.price AS delivery_price, so.order_total, soi.id, soi.product_item_id, soi.order_id, soi.size, soi.color, soi.quantity, soi.price, soi.created_at, soi.updated_at, p.name AS product_name,
pimg.product_image_1 AS product_image,
pi.active AS product_active, a.address_line, a.region, a.city,
DENSE_RANK() OVER(ORDER BY so.id) as order_number, pt.value AS payment_type 
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
LEFT JOIN "product_item" AS pi ON pi.id = soi.product_item_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "address" AS a ON a.id = so.shipping_address_id
LEFT JOIN "payment_method" AS pm ON pm.id = so.payment_method_id
LEFT JOIN "payment_type" AS pt ON pt.id = pm.payment_type_id
LEFT JOIN "shipping_method" AS sm ON sm.id = so.shipping_method_id
WHERE so.user_id = $1
AND soi.order_id = $2
`

type ListShopOrderItemsByUserIDOrderIDParams struct {
	UserID  int64 `json:"user_id"`
	OrderID int64 `json:"order_id"`
}

type ListShopOrderItemsByUserIDOrderIDRow struct {
	Status        null.String `json:"status"`
	TrackNumber   null.String `json:"track_number"`
	DeliveryPrice null.String `json:"delivery_price"`
	OrderTotal    null.String `json:"order_total"`
	ID            int64       `json:"id"`
	ProductItemID int64       `json:"product_item_id"`
	OrderID       int64       `json:"order_id"`
	Size          string      `json:"size"`
	Color         string      `json:"color"`
	Quantity      int32       `json:"quantity"`
	Price         string      `json:"price"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	ProductName   null.String `json:"product_name"`
	ProductImage  null.String `json:"product_image"`
	ProductActive null.Bool   `json:"product_active"`
	AddressLine   null.String `json:"address_line"`
	Region        null.String `json:"region"`
	City          null.String `json:"city"`
	OrderNumber   int64       `json:"order_number"`
	PaymentType   null.String `json:"payment_type"`
}

// SELECT * FROM "shop_order_item"
// WHERE order_id = $1
// ORDER BY id;
// pi.product_image,
func (q *Queries) ListShopOrderItemsByUserIDOrderID(ctx context.Context, arg ListShopOrderItemsByUserIDOrderIDParams) ([]ListShopOrderItemsByUserIDOrderIDRow, error) {
	rows, err := q.db.Query(ctx, listShopOrderItemsByUserIDOrderID, arg.UserID, arg.OrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListShopOrderItemsByUserIDOrderIDRow{}
	for rows.Next() {
		var i ListShopOrderItemsByUserIDOrderIDRow
		if err := rows.Scan(
			&i.Status,
			&i.TrackNumber,
			&i.DeliveryPrice,
			&i.OrderTotal,
			&i.ID,
			&i.ProductItemID,
			&i.OrderID,
			&i.Size,
			&i.Color,
			&i.Quantity,
			&i.Price,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ProductName,
			&i.ProductImage,
			&i.ProductActive,
			&i.AddressLine,
			&i.Region,
			&i.City,
			&i.OrderNumber,
			&i.PaymentType,
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

const updateShopOrderItem = `-- name: UpdateShopOrderItem :one

UPDATE "shop_order_item"
SET 
product_item_id = COALESCE($1,product_item_id),
order_id = COALESCE($2,order_id),
quantity = COALESCE($3,quantity),
price = COALESCE($4,price),
updated_at = now()
WHERE id = $5
RETURNING id, product_item_id, order_id, size, color, quantity, price, created_at, updated_at
`

type UpdateShopOrderItemParams struct {
	ProductItemID null.Int    `json:"product_item_id"`
	OrderID       null.Int    `json:"order_id"`
	Quantity      null.Int    `json:"quantity"`
	Price         null.String `json:"price"`
	ID            int64       `json:"id"`
}

// -- name: ListShopOrderItemsByOrderID :many
// SELECT * FROM "shop_order_item"
// WHERE order_id = $1
// ORDER BY id;
func (q *Queries) UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, updateShopOrderItem,
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
		&i.Size,
		&i.Color,
		&i.Quantity,
		&i.Price,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
