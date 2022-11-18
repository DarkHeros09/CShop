// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: category_promotion.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createCategoryPromotion = `-- name: CreateCategoryPromotion :one
INSERT INTO "category_promotion" (
  category_id,
  promotion_id,
  active
) VALUES (
  $1, $2, $3
)
RETURNING category_id, promotion_id, active
`

type CreateCategoryPromotionParams struct {
	CategoryID  int64 `json:"category_id"`
	PromotionID int64 `json:"promotion_id"`
	Active      bool  `json:"active"`
}

func (q *Queries) CreateCategoryPromotion(ctx context.Context, arg CreateCategoryPromotionParams) (CategoryPromotion, error) {
	row := q.db.QueryRow(ctx, createCategoryPromotion, arg.CategoryID, arg.PromotionID, arg.Active)
	var i CategoryPromotion
	err := row.Scan(&i.CategoryID, &i.PromotionID, &i.Active)
	return i, err
}

const deleteCategoryPromotion = `-- name: DeleteCategoryPromotion :exec
DELETE FROM "category_promotion"
WHERE category_id = $1
`

func (q *Queries) DeleteCategoryPromotion(ctx context.Context, categoryID int64) error {
	_, err := q.db.Exec(ctx, deleteCategoryPromotion, categoryID)
	return err
}

const getCategoryPromotion = `-- name: GetCategoryPromotion :one
SELECT category_id, promotion_id, active FROM "category_promotion"
WHERE category_id = $1 LIMIT 1
`

func (q *Queries) GetCategoryPromotion(ctx context.Context, categoryID int64) (CategoryPromotion, error) {
	row := q.db.QueryRow(ctx, getCategoryPromotion, categoryID)
	var i CategoryPromotion
	err := row.Scan(&i.CategoryID, &i.PromotionID, &i.Active)
	return i, err
}

const listCategoryPromotions = `-- name: ListCategoryPromotions :many
SELECT category_id, promotion_id, active FROM "category_promotion"
ORDER BY category_id
LIMIT $1
OFFSET $2
`

type ListCategoryPromotionsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListCategoryPromotions(ctx context.Context, arg ListCategoryPromotionsParams) ([]CategoryPromotion, error) {
	rows, err := q.db.Query(ctx, listCategoryPromotions, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []CategoryPromotion{}
	for rows.Next() {
		var i CategoryPromotion
		if err := rows.Scan(&i.CategoryID, &i.PromotionID, &i.Active); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCategoryPromotion = `-- name: UpdateCategoryPromotion :one
UPDATE "category_promotion"
SET
promotion_id = COALESCE($1,promotion_id),
active = COALESCE($2,active)
WHERE category_id = $3
RETURNING category_id, promotion_id, active
`

type UpdateCategoryPromotionParams struct {
	PromotionID null.Int  `json:"promotion_id"`
	Active      null.Bool `json:"active"`
	CategoryID  int64     `json:"category_id"`
}

func (q *Queries) UpdateCategoryPromotion(ctx context.Context, arg UpdateCategoryPromotionParams) (CategoryPromotion, error) {
	row := q.db.QueryRow(ctx, updateCategoryPromotion, arg.PromotionID, arg.Active, arg.CategoryID)
	var i CategoryPromotion
	err := row.Scan(&i.CategoryID, &i.PromotionID, &i.Active)
	return i, err
}
