// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: user_review.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const createUserReview = `-- name: CreateUserReview :one
INSERT INTO "user_review" (
  user_id,
  ordered_product_id,
  rating_value
) VALUES (
  $1, $2, $3
)
RETURNING id, user_id, ordered_product_id, rating_value, created_at, updated_at
`

type CreateUserReviewParams struct {
	UserID           int64 `json:"user_id"`
	OrderedProductID int64 `json:"ordered_product_id"`
	RatingValue      int32 `json:"rating_value"`
}

func (q *Queries) CreateUserReview(ctx context.Context, arg CreateUserReviewParams) (UserReview, error) {
	row := q.db.QueryRow(ctx, createUserReview, arg.UserID, arg.OrderedProductID, arg.RatingValue)
	var i UserReview
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.OrderedProductID,
		&i.RatingValue,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUserReview = `-- name: DeleteUserReview :one
DELETE FROM "user_review"
WHERE id = $1
And user_id =$2
RETURNING id, user_id, ordered_product_id, rating_value, created_at, updated_at
`

type DeleteUserReviewParams struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) DeleteUserReview(ctx context.Context, arg DeleteUserReviewParams) (UserReview, error) {
	row := q.db.QueryRow(ctx, deleteUserReview, arg.ID, arg.UserID)
	var i UserReview
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.OrderedProductID,
		&i.RatingValue,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserReview = `-- name: GetUserReview :one
SELECT id, user_id, ordered_product_id, rating_value, created_at, updated_at FROM "user_review"
WHERE id = $1 
AND user_id = $2
LIMIT 1
`

type GetUserReviewParams struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) GetUserReview(ctx context.Context, arg GetUserReviewParams) (UserReview, error) {
	row := q.db.QueryRow(ctx, getUserReview, arg.ID, arg.UserID)
	var i UserReview
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.OrderedProductID,
		&i.RatingValue,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUserReviews = `-- name: ListUserReviews :many
SELECT id, user_id, ordered_product_id, rating_value, created_at, updated_at FROM "user_review"
WHERE user_id = $3
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListUserReviewsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) ListUserReviews(ctx context.Context, arg ListUserReviewsParams) ([]UserReview, error) {
	rows, err := q.db.Query(ctx, listUserReviews, arg.Limit, arg.Offset, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserReview{}
	for rows.Next() {
		var i UserReview
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.OrderedProductID,
			&i.RatingValue,
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

const updateUserReview = `-- name: UpdateUserReview :one
UPDATE "user_review"
SET 
ordered_product_id = COALESCE($1,ordered_product_id),
rating_value = COALESCE($2,rating_value),
updated_at = now()
WHERE id = $3
AND user_id = $4
RETURNING id, user_id, ordered_product_id, rating_value, created_at, updated_at
`

type UpdateUserReviewParams struct {
	OrderedProductID null.Int `json:"ordered_product_id"`
	RatingValue      null.Int `json:"rating_value"`
	ID               int64    `json:"id"`
	UserID           int64    `json:"user_id"`
}

func (q *Queries) UpdateUserReview(ctx context.Context, arg UpdateUserReviewParams) (UserReview, error) {
	row := q.db.QueryRow(ctx, updateUserReview,
		arg.OrderedProductID,
		arg.RatingValue,
		arg.ID,
		arg.UserID,
	)
	var i UserReview
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.OrderedProductID,
		&i.RatingValue,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
