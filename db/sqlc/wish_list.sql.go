// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: wish_list.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createWishList = `-- name: CreateWishList :one
INSERT INTO "wish_list" (
  user_id
) VALUES (
  $1
)
RETURNING id, user_id, created_at, updated_at
`

func (q *Queries) CreateWishList(ctx context.Context, userID int64) (WishList, error) {
	row := q.db.QueryRow(ctx, createWishList, userID)
	var i WishList
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteWishList = `-- name: DeleteWishList :exec
DELETE FROM "wish_list"
WHERE id = $1
`

func (q *Queries) DeleteWishList(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteWishList, id)
	return err
}

const getWishList = `-- name: GetWishList :one
SELECT id, user_id, created_at, updated_at FROM "wish_list"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetWishList(ctx context.Context, id int64) (WishList, error) {
	row := q.db.QueryRow(ctx, getWishList, id)
	var i WishList
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getWishListByUserID = `-- name: GetWishListByUserID :one
SELECT id, user_id, created_at, updated_at FROM "wish_list"
WHERE user_id = $1 LIMIT 1
`

func (q *Queries) GetWishListByUserID(ctx context.Context, userID int64) (WishList, error) {
	row := q.db.QueryRow(ctx, getWishListByUserID, userID)
	var i WishList
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listWishLists = `-- name: ListWishLists :many
SELECT id, user_id, created_at, updated_at FROM "wish_list"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListWishListsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListWishLists(ctx context.Context, arg ListWishListsParams) ([]WishList, error) {
	rows, err := q.db.Query(ctx, listWishLists, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishList{}
	for rows.Next() {
		var i WishList
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

const updateWishList = `-- name: UpdateWishList :one
UPDATE "wish_list"
SET 
user_id = COALESCE($1,user_id),
updated_at = now()
WHERE id = $2
RETURNING id, user_id, created_at, updated_at
`

type UpdateWishListParams struct {
	UserID null.Int `json:"user_id"`
	ID     int64    `json:"id"`
}

func (q *Queries) UpdateWishList(ctx context.Context, arg UpdateWishListParams) (WishList, error) {
	row := q.db.QueryRow(ctx, updateWishList, arg.UserID, arg.ID)
	var i WishList
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
