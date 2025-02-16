// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: shop_order_item.sql

package db

import (
	"context"
	"time"

	null "github.com/guregu/null/v5"
)

const createShopOrderItem = `-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  quantity,
  price,
  discount,
  shipping_method_price
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, product_item_id, order_id, quantity, price, discount, shipping_method_price, created_at, updated_at
`

type CreateShopOrderItemParams struct {
	ProductItemID       int64  `json:"product_item_id"`
	OrderID             int64  `json:"order_id"`
	Quantity            int32  `json:"quantity"`
	Price               string `json:"price"`
	Discount            int32  `json:"discount"`
	ShippingMethodPrice string `json:"shipping_method_price"`
}

func (q *Queries) CreateShopOrderItem(ctx context.Context, arg CreateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, createShopOrderItem,
		arg.ProductItemID,
		arg.OrderID,
		arg.Quantity,
		arg.Price,
		arg.Discount,
		arg.ShippingMethodPrice,
	)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.Discount,
		&i.ShippingMethodPrice,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteShopOrderItem = `-- name: DeleteShopOrderItem :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $2
    AND active = TRUE
    )
DELETE FROM "shop_order_item"
WHERE "shop_order_item".id = $1
AND (SELECT is_admin FROM t1) = 1
RETURNING id, product_item_id, order_id, quantity, price, discount, shipping_method_price, created_at, updated_at
`

type DeleteShopOrderItemParams struct {
	ID      int64 `json:"id"`
	AdminID int64 `json:"admin_id"`
}

func (q *Queries) DeleteShopOrderItem(ctx context.Context, arg DeleteShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, deleteShopOrderItem, arg.ID, arg.AdminID)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.Discount,
		&i.ShippingMethodPrice,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getShopOrderItem = `-- name: GetShopOrderItem :one
SELECT id, product_item_id, order_id, quantity, price, discount, shipping_method_price, created_at, updated_at FROM "shop_order_item"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, getShopOrderItem, id)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.Discount,
		&i.ShippingMethodPrice,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getShopOrderItemByUserIDOrderID = `-- name: GetShopOrderItemByUserIDOrderID :one
SELECT soi.id, soi.product_item_id, soi.order_id, soi.quantity, soi.price, soi.discount, soi.shipping_method_price, soi.created_at, soi.updated_at, so.user_id
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
	ID                  int64     `json:"id"`
	ProductItemID       int64     `json:"product_item_id"`
	OrderID             int64     `json:"order_id"`
	Quantity            int32     `json:"quantity"`
	Price               string    `json:"price"`
	Discount            int32     `json:"discount"`
	ShippingMethodPrice string    `json:"shipping_method_price"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	UserID              null.Int  `json:"user_id"`
}

func (q *Queries) GetShopOrderItemByUserIDOrderID(ctx context.Context, arg GetShopOrderItemByUserIDOrderIDParams) (GetShopOrderItemByUserIDOrderIDRow, error) {
	row := q.db.QueryRow(ctx, getShopOrderItemByUserIDOrderID, arg.UserID, arg.OrderID)
	var i GetShopOrderItemByUserIDOrderIDRow
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.Discount,
		&i.ShippingMethodPrice,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}

const listShopOrderItems = `-- name: ListShopOrderItems :many
SELECT id, product_item_id, order_id, quantity, price, discount, shipping_method_price, created_at, updated_at FROM "shop_order_item"
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
			&i.Quantity,
			&i.Price,
			&i.Discount,
			&i.ShippingMethodPrice,
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

SELECT so.id, so.track_number, so.order_number, so.user_id, so.payment_type_id, so.shipping_address_id, so.order_total, so.shipping_method_id, so.order_status_id, so.created_at, so.updated_at, so.completed_at, soi.id, soi.product_item_id, soi.order_id, soi.quantity, soi.price, soi.discount, soi.shipping_method_price, soi.created_at, soi.updated_at 
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
	ID                  int64       `json:"id"`
	TrackNumber         string      `json:"track_number"`
	OrderNumber         int32       `json:"order_number"`
	UserID              int64       `json:"user_id"`
	PaymentTypeID       int64       `json:"payment_type_id"`
	ShippingAddressID   int64       `json:"shipping_address_id"`
	OrderTotal          string      `json:"order_total"`
	ShippingMethodID    int64       `json:"shipping_method_id"`
	OrderStatusID       null.Int    `json:"order_status_id"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	CompletedAt         time.Time   `json:"completed_at"`
	ID_2                null.Int    `json:"id_2"`
	ProductItemID       null.Int    `json:"product_item_id"`
	OrderID             null.Int    `json:"order_id"`
	Quantity            null.Int    `json:"quantity"`
	Price               null.String `json:"price"`
	Discount            null.Int    `json:"discount"`
	ShippingMethodPrice null.String `json:"shipping_method_price"`
	CreatedAt_2         null.Time   `json:"created_at_2"`
	UpdatedAt_2         null.Time   `json:"updated_at_2"`
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
			&i.OrderNumber,
			&i.UserID,
			&i.PaymentTypeID,
			&i.ShippingAddressID,
			&i.OrderTotal,
			&i.ShippingMethodID,
			&i.OrderStatusID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CompletedAt,
			&i.ID_2,
			&i.ProductItemID,
			&i.OrderID,
			&i.Quantity,
			&i.Price,
			&i.Discount,
			&i.ShippingMethodPrice,
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
SELECT os.status, so.track_number, soi.shipping_method_price AS delivery_price, so.order_total, soi.id, soi.product_item_id, soi.order_id, soi.quantity, soi.price, soi.discount, soi.shipping_method_price, soi.created_at, soi.updated_at, p.name AS product_name, pt.value as payment_type,
pimg.product_image_1 AS product_image,
pcolor.color_value AS product_color, psize.size_value AS product_size,
pi.active AS product_active, a.address_line, a.region, a.city,
DENSE_RANK() OVER(ORDER BY so.id) as order_number
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
LEFT JOIN "product_item" AS pi ON pi.id = soi.product_item_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_size" AS psize ON psize.product_item_id = pi.id
LEFT JOIN "product_color" AS pcolor ON pcolor.id = pi.color_id
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "address" AS a ON a.id = so.shipping_address_id
LEFT JOIN "payment_type" AS pt ON pt.id = so.payment_type_id
WHERE so.user_id = $1
AND soi.order_id = $2
`

type ListShopOrderItemsByUserIDOrderIDParams struct {
	UserID  int64 `json:"user_id"`
	OrderID int64 `json:"order_id"`
}

type ListShopOrderItemsByUserIDOrderIDRow struct {
	Status              null.String `json:"status"`
	TrackNumber         null.String `json:"track_number"`
	DeliveryPrice       string      `json:"delivery_price"`
	OrderTotal          null.String `json:"order_total"`
	ID                  int64       `json:"id"`
	ProductItemID       int64       `json:"product_item_id"`
	OrderID             int64       `json:"order_id"`
	Quantity            int32       `json:"quantity"`
	Price               string      `json:"price"`
	Discount            int32       `json:"discount"`
	ShippingMethodPrice string      `json:"shipping_method_price"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	ProductName         null.String `json:"product_name"`
	PaymentType         null.String `json:"payment_type"`
	ProductImage        null.String `json:"product_image"`
	ProductColor        null.String `json:"product_color"`
	ProductSize         null.String `json:"product_size"`
	ProductActive       null.Bool   `json:"product_active"`
	AddressLine         null.String `json:"address_line"`
	Region              null.String `json:"region"`
	City                null.String `json:"city"`
	OrderNumber         int64       `json:"order_number"`
}

// SELECT * FROM "shop_order_item"
// WHERE order_id = $1
// ORDER BY id;
// pi.product_image,
// , pt.value AS payment_type
// LEFT JOIN "payment_method" AS pm ON pm.id = so.payment_method_id
// LEFT JOIN "shipping_method" AS sm ON sm.id = so.shipping_method_id
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
			&i.Quantity,
			&i.Price,
			&i.Discount,
			&i.ShippingMethodPrice,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ProductName,
			&i.PaymentType,
			&i.ProductImage,
			&i.ProductColor,
			&i.ProductSize,
			&i.ProductActive,
			&i.AddressLine,
			&i.Region,
			&i.City,
			&i.OrderNumber,
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
quantity = COALESCE($1,quantity),
price = COALESCE($2,price),
discount = COALESCE($3,discount),
shipping_method_price = COALESCE($4,shipping_method_price),
updated_at = now()
WHERE id = $5
AND order_id = $6
AND product_item_id = $7
RETURNING id, product_item_id, order_id, quantity, price, discount, shipping_method_price, created_at, updated_at
`

type UpdateShopOrderItemParams struct {
	Quantity            null.Int    `json:"quantity"`
	Price               null.String `json:"price"`
	Discount            null.Int    `json:"discount"`
	ShippingMethodPrice null.String `json:"shipping_method_price"`
	ID                  int64       `json:"id"`
	OrderID             int64       `json:"order_id"`
	ProductItemID       int64       `json:"product_item_id"`
}

// -- name: ListShopOrderItemsByOrderID :many
// SELECT * FROM "shop_order_item"
// WHERE order_id = $1
// ORDER BY id;
func (q *Queries) UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error) {
	row := q.db.QueryRow(ctx, updateShopOrderItem,
		arg.Quantity,
		arg.Price,
		arg.Discount,
		arg.ShippingMethodPrice,
		arg.ID,
		arg.OrderID,
		arg.ProductItemID,
	)
	var i ShopOrderItem
	err := row.Scan(
		&i.ID,
		&i.ProductItemID,
		&i.OrderID,
		&i.Quantity,
		&i.Price,
		&i.Discount,
		&i.ShippingMethodPrice,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
