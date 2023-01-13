// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: wish_list_item.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createWishListItem = `-- name: CreateWishListItem :one
INSERT INTO "wish_list_item" (
  wish_list_id,
  product_item_id
) VALUES (
  $1, $2
)
RETURNING id, wish_list_id, product_item_id, created_at, updated_at
`

type CreateWishListItemParams struct {
	WishListID    int64 `json:"wish_list_id"`
	ProductItemID int64 `json:"product_item_id"`
}

func (q *Queries) CreateWishListItem(ctx context.Context, arg CreateWishListItemParams) (WishListItem, error) {
	row := q.db.QueryRow(ctx, createWishListItem, arg.WishListID, arg.ProductItemID)
	var i WishListItem
	err := row.Scan(
		&i.ID,
		&i.WishListID,
		&i.ProductItemID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteWishListItem = `-- name: DeleteWishListItem :exec
DELETE FROM "wish_list_item" AS wli
WHERE wli.id = $1
AND wli.wish_list_id = $2
`

type DeleteWishListItemParams struct {
	ID         int64 `json:"id"`
	WishListID int64 `json:"wish_list_id"`
}

// WITH t1 AS (
//
//	SELECT id FROM "wish_list" AS wl
//	WHERE wl.user_id = sqlc.arg(user_id)
//
// )
func (q *Queries) DeleteWishListItem(ctx context.Context, arg DeleteWishListItemParams) error {
	_, err := q.db.Exec(ctx, deleteWishListItem, arg.ID, arg.WishListID)
	return err
}

const deleteWishListItemAll = `-- name: DeleteWishListItemAll :many
DELETE FROM "wish_list_item"
WHERE wish_list_id = $1
RETURNING id, wish_list_id, product_item_id, created_at, updated_at
`

// WITH t1 AS(
//
//	SELECT id FROM "wish_list" WHERE user_id = $1
//
// )
func (q *Queries) DeleteWishListItemAll(ctx context.Context, wishListID int64) ([]WishListItem, error) {
	rows, err := q.db.Query(ctx, deleteWishListItemAll, wishListID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishListItem{}
	for rows.Next() {
		var i WishListItem
		if err := rows.Scan(
			&i.ID,
			&i.WishListID,
			&i.ProductItemID,
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

const getWishListItem = `-- name: GetWishListItem :one
SELECT id, wish_list_id, product_item_id, created_at, updated_at FROM "wish_list_item"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetWishListItem(ctx context.Context, id int64) (WishListItem, error) {
	row := q.db.QueryRow(ctx, getWishListItem, id)
	var i WishListItem
	err := row.Scan(
		&i.ID,
		&i.WishListID,
		&i.ProductItemID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getWishListItemByUserIDCartID = `-- name: GetWishListItemByUserIDCartID :one
SELECT wli.id, wli.wish_list_id, wli.product_item_id, wli.created_at, wli.updated_at
FROM "wish_list_item" AS wli
LEFT JOIN "wish_list" AS wl ON wl.id = wli.wish_list_id
WHERE wl.user_id = $1
AND wli.id = $2
AND wli.wish_list_id = $3
LIMIT 1
`

type GetWishListItemByUserIDCartIDParams struct {
	UserID     int64 `json:"user_id"`
	ID         int64 `json:"id"`
	WishListID int64 `json:"wish_list_id"`
}

func (q *Queries) GetWishListItemByUserIDCartID(ctx context.Context, arg GetWishListItemByUserIDCartIDParams) (WishListItem, error) {
	row := q.db.QueryRow(ctx, getWishListItemByUserIDCartID, arg.UserID, arg.ID, arg.WishListID)
	var i WishListItem
	err := row.Scan(
		&i.ID,
		&i.WishListID,
		&i.ProductItemID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listWishListItems = `-- name: ListWishListItems :many
SELECT id, wish_list_id, product_item_id, created_at, updated_at FROM "wish_list_item"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListWishListItemsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListWishListItems(ctx context.Context, arg ListWishListItemsParams) ([]WishListItem, error) {
	rows, err := q.db.Query(ctx, listWishListItems, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishListItem{}
	for rows.Next() {
		var i WishListItem
		if err := rows.Scan(
			&i.ID,
			&i.WishListID,
			&i.ProductItemID,
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

const listWishListItemsByCartID = `-- name: ListWishListItemsByCartID :many
SELECT id, wish_list_id, product_item_id, created_at, updated_at FROM "wish_list_item"
WHERE wish_list_id = $1
ORDER BY id
`

func (q *Queries) ListWishListItemsByCartID(ctx context.Context, wishListID int64) ([]WishListItem, error) {
	rows, err := q.db.Query(ctx, listWishListItemsByCartID, wishListID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishListItem{}
	for rows.Next() {
		var i WishListItem
		if err := rows.Scan(
			&i.ID,
			&i.WishListID,
			&i.ProductItemID,
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

const listWishListItemsByUserID = `-- name: ListWishListItemsByUserID :many
SELECT wl.user_id, wli.id, wli.wish_list_id, wli.product_item_id, wli.created_at, wli.updated_at
FROM "wish_list" AS wl
LEFT JOIN "wish_list_item" AS wli ON wli.wish_list_id = wl.id
WHERE wl.user_id = $1
`

type ListWishListItemsByUserIDRow struct {
	UserID        int64     `json:"user_id"`
	ID            null.Int  `json:"id"`
	WishListID    null.Int  `json:"wish_list_id"`
	ProductItemID null.Int  `json:"product_item_id"`
	CreatedAt     null.Time `json:"created_at"`
	UpdatedAt     null.Time `json:"updated_at"`
}

func (q *Queries) ListWishListItemsByUserID(ctx context.Context, userID int64) ([]ListWishListItemsByUserIDRow, error) {
	rows, err := q.db.Query(ctx, listWishListItemsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListWishListItemsByUserIDRow{}
	for rows.Next() {
		var i ListWishListItemsByUserIDRow
		if err := rows.Scan(
			&i.UserID,
			&i.ID,
			&i.WishListID,
			&i.ProductItemID,
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

const updateWishListItem = `-- name: UpdateWishListItem :one

UPDATE "wish_list_item" AS wli
SET 
product_item_id = COALESCE($1,product_item_id)
WHERE wli.id = $2
AND wli.wish_list_id = $3
RETURNING id, wish_list_id, product_item_id, created_at, updated_at
`

type UpdateWishListItemParams struct {
	ProductItemID null.Int `json:"product_item_id"`
	ID            int64    `json:"id"`
	WishListID    int64    `json:"wish_list_id"`
}

// WITH t1 AS (
//
//	SELECT user_id FROM "wish_list" AS wl
//	WHERE wl.id = sqlc.arg(wish_list_id)
//
// )
func (q *Queries) UpdateWishListItem(ctx context.Context, arg UpdateWishListItemParams) (WishListItem, error) {
	row := q.db.QueryRow(ctx, updateWishListItem, arg.ProductItemID, arg.ID, arg.WishListID)
	var i WishListItem
	err := row.Scan(
		&i.ID,
		&i.WishListID,
		&i.ProductItemID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
