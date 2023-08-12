// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
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
  product_promotion_image,
  active
) VALUES (
  $1, $2, $3, $4
)
RETURNING product_id, promotion_id, product_promotion_image, active
`

type CreateProductPromotionParams struct {
	ProductID             int64       `json:"product_id"`
	PromotionID           int64       `json:"promotion_id"`
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                bool        `json:"active"`
}

func (q *Queries) CreateProductPromotion(ctx context.Context, arg CreateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, createProductPromotion,
		arg.ProductID,
		arg.PromotionID,
		arg.ProductPromotionImage,
		arg.Active,
	)
	var i ProductPromotion
	err := row.Scan(
		&i.ProductID,
		&i.PromotionID,
		&i.ProductPromotionImage,
		&i.Active,
	)
	return i, err
}

const deleteProductPromotion = `-- name: DeleteProductPromotion :exec
DELETE FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
RETURNING product_id, promotion_id, product_promotion_image, active
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
SELECT product_id, promotion_id, product_promotion_image, active FROM "product_promotion"
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
	err := row.Scan(
		&i.ProductID,
		&i.PromotionID,
		&i.ProductPromotionImage,
		&i.Active,
	)
	return i, err
}

const listProductPromotions = `-- name: ListProductPromotions :many
SELECT product_id, promotion_id, product_promotion_image, active FROM "product_promotion"
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
		if err := rows.Scan(
			&i.ProductID,
			&i.PromotionID,
			&i.ProductPromotionImage,
			&i.Active,
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

const listProductPromotionsWithImages = `-- name: ListProductPromotionsWithImages :many
SELECT product_id, promotion_id, product_promotion_image, pp.active, id, category_id, brand_id, name, description, p.active, created_at, updated_at, search FROM "product_promotion" AS pp
LEFT JOIN "product" AS p ON p.id = pp.product_id
WHERE pp.product_promotion_image IS NOT NULL AND pp.active = true
`

type ListProductPromotionsWithImagesRow struct {
	ProductID             int64       `json:"product_id"`
	PromotionID           int64       `json:"promotion_id"`
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                bool        `json:"active"`
	ID                    null.Int    `json:"id"`
	CategoryID            null.Int    `json:"category_id"`
	BrandID               null.Int    `json:"brand_id"`
	Name                  null.String `json:"name"`
	Description           null.String `json:"description"`
	Active_2              null.Bool   `json:"active_2"`
	CreatedAt             null.Time   `json:"created_at"`
	UpdatedAt             null.Time   `json:"updated_at"`
	Search                null.String `json:"search"`
}

func (q *Queries) ListProductPromotionsWithImages(ctx context.Context) ([]ListProductPromotionsWithImagesRow, error) {
	rows, err := q.db.Query(ctx, listProductPromotionsWithImages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductPromotionsWithImagesRow{}
	for rows.Next() {
		var i ListProductPromotionsWithImagesRow
		if err := rows.Scan(
			&i.ProductID,
			&i.PromotionID,
			&i.ProductPromotionImage,
			&i.Active,
			&i.ID,
			&i.CategoryID,
			&i.BrandID,
			&i.Name,
			&i.Description,
			&i.Active_2,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Search,
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

const updateProductPromotion = `-- name: UpdateProductPromotion :one
UPDATE "product_promotion"
SET
product_promotion_image = COALESCE($1,product_promotion_image),
active = COALESCE($2,active)
WHERE product_id = $3
AND promotion_id = $4
RETURNING product_id, promotion_id, product_promotion_image, active
`

type UpdateProductPromotionParams struct {
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                null.Bool   `json:"active"`
	ProductID             int64       `json:"product_id"`
	PromotionID           int64       `json:"promotion_id"`
}

func (q *Queries) UpdateProductPromotion(ctx context.Context, arg UpdateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, updateProductPromotion,
		arg.ProductPromotionImage,
		arg.Active,
		arg.ProductID,
		arg.PromotionID,
	)
	var i ProductPromotion
	err := row.Scan(
		&i.ProductID,
		&i.PromotionID,
		&i.ProductPromotionImage,
		&i.Active,
	)
	return i, err
}
