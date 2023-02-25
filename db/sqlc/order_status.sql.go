// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: order_status.sql

package db

import (
	"context"
	"time"

	"github.com/guregu/null"
)

const createOrderStatus = `-- name: CreateOrderStatus :one
INSERT INTO "order_status" (
  status
) VALUES (
  $1
)
RETURNING id, status, created_at, updated_at
`

func (q *Queries) CreateOrderStatus(ctx context.Context, status string) (OrderStatus, error) {
	row := q.db.QueryRow(ctx, createOrderStatus, status)
	var i OrderStatus
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteOrderStatus = `-- name: DeleteOrderStatus :exec
DELETE FROM "order_status"
WHERE id = $1
`

func (q *Queries) DeleteOrderStatus(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteOrderStatus, id)
	return err
}

const getOrderStatus = `-- name: GetOrderStatus :one
SELECT id, status, created_at, updated_at FROM "order_status"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetOrderStatus(ctx context.Context, id int64) (OrderStatus, error) {
	row := q.db.QueryRow(ctx, getOrderStatus, id)
	var i OrderStatus
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderStatusByUserID = `-- name: GetOrderStatusByUserID :one
SELECT os.id, os.status, os.created_at, os.updated_at, so.user_id
FROM "order_status" AS os
LEFT JOIN "shop_order" AS so ON so.order_status_id = os.id
WHERE so.user_id = $1
AND os.id = $2
LIMIT 1
`

type GetOrderStatusByUserIDParams struct {
	UserID int64 `json:"user_id"`
	ID     int64 `json:"id"`
}

type GetOrderStatusByUserIDRow struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    null.Int  `json:"user_id"`
}

func (q *Queries) GetOrderStatusByUserID(ctx context.Context, arg GetOrderStatusByUserIDParams) (GetOrderStatusByUserIDRow, error) {
	row := q.db.QueryRow(ctx, getOrderStatusByUserID, arg.UserID, arg.ID)
	var i GetOrderStatusByUserIDRow
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
	)
	return i, err
}

const listOrderStatuses = `-- name: ListOrderStatuses :many
SELECT id, status, created_at, updated_at FROM "order_status"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListOrderStatusesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListOrderStatuses(ctx context.Context, arg ListOrderStatusesParams) ([]OrderStatus, error) {
	rows, err := q.db.Query(ctx, listOrderStatuses, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OrderStatus{}
	for rows.Next() {
		var i OrderStatus
		if err := rows.Scan(
			&i.ID,
			&i.Status,
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

const listOrderStatusesByUserID = `-- name: ListOrderStatusesByUserID :many
SELECT os.id, os.status, os.created_at, os.updated_at, so.user_id
FROM "order_status" AS os
LEFT JOIN "shop_order" AS so ON so.order_status_id = os.id
WHERE so.user_id = $3
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListOrderStatusesByUserIDParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserID int64 `json:"user_id"`
}

type ListOrderStatusesByUserIDRow struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    null.Int  `json:"user_id"`
}

func (q *Queries) ListOrderStatusesByUserID(ctx context.Context, arg ListOrderStatusesByUserIDParams) ([]ListOrderStatusesByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listOrderStatusesByUserID, arg.Limit, arg.Offset, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListOrderStatusesByUserIDRow{}
	for rows.Next() {
		var i ListOrderStatusesByUserIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Status,
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

const updateOrderStatus = `-- name: UpdateOrderStatus :one
UPDATE "order_status"
SET 
status = COALESCE($1,status),
updated_at = now()
WHERE id = $2
RETURNING id, status, created_at, updated_at
`

type UpdateOrderStatusParams struct {
	Status null.String `json:"status"`
	ID     int64       `json:"id"`
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (OrderStatus, error) {
	row := q.db.QueryRow(ctx, updateOrderStatus, arg.Status, arg.ID)
	var i OrderStatus
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
