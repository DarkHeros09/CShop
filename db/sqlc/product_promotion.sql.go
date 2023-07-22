// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: product_promotion.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createProductPromotion = `-- name: CreateProductPromotion :one
INSERT INTO "product_promotion" (
  product_id,
  promotion_id,
  active
) VALUES (
  $1, $2, $3
)
RETURNING product_id, promotion_id, active
`

type CreateProductPromotionParams struct {
	ProductID   int64 `json:"product_id"`
	PromotionID int64 `json:"promotion_id"`
	Active      bool  `json:"active"`
}

func (q *Queries) CreateProductPromotion(ctx context.Context, arg CreateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, createProductPromotion, arg.ProductID, arg.PromotionID, arg.Active)
	var i ProductPromotion
	err := row.Scan(&i.ProductID, &i.PromotionID, &i.Active)
	return i, err
}

const deleteProductPromotion = `-- name: DeleteProductPromotion :exec
DELETE FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
RETURNING product_id, promotion_id, active
`

type DeleteProductPromotionParams struct {
	ProductID   int64 `json:"product_id"`
	PromotionID int64 `json:"promotion_id"`
}

func (q *Queries) DeleteProductPromotion(ctx context.Context, arg DeleteProductPromotionParams) error {
	_, err := q.db.Exec(ctx, deleteProductPromotion, arg.ProductID, arg.PromotionID)
	return err
}

const getProductPromotion = `-- name: GetProductPromotion :one
SELECT product_id, promotion_id, active FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
LIMIT 1
`

type GetProductPromotionParams struct {
	ProductID   int64 `json:"product_id"`
	PromotionID int64 `json:"promotion_id"`
}

func (q *Queries) GetProductPromotion(ctx context.Context, arg GetProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, getProductPromotion, arg.ProductID, arg.PromotionID)
	var i ProductPromotion
	err := row.Scan(&i.ProductID, &i.PromotionID, &i.Active)
	return i, err
}

const listProductPromotions = `-- name: ListProductPromotions :many
SELECT product_id, promotion_id, active FROM "product_promotion"
ORDER BY product_id
LIMIT $1
OFFSET $2
`

type ListProductPromotionsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListProductPromotions(ctx context.Context, arg ListProductPromotionsParams) ([]ProductPromotion, error) {
	rows, err := q.db.Query(ctx, listProductPromotions, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProductPromotion{}
	for rows.Next() {
		var i ProductPromotion
		if err := rows.Scan(&i.ProductID, &i.PromotionID, &i.Active); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateProductPromotion = `-- name: UpdateProductPromotion :one
UPDATE "product_promotion"
SET
active = COALESCE($1,active)
WHERE product_id = $2
AND promotion_id = $3
RETURNING product_id, promotion_id, active
`

type UpdateProductPromotionParams struct {
	Active      null.Bool `json:"active"`
	ProductID   int64     `json:"product_id"`
	PromotionID int64     `json:"promotion_id"`
}

func (q *Queries) UpdateProductPromotion(ctx context.Context, arg UpdateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, updateProductPromotion, arg.Active, arg.ProductID, arg.PromotionID)
	var i ProductPromotion
	err := row.Scan(&i.ProductID, &i.PromotionID, &i.Active)
	return i, err
}
