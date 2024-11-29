// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: product_promotion.sql

package db

import (
	"context"
	"time"

	null "github.com/guregu/null/v5"
)

const adminCreateProductPromotion = `-- name: AdminCreateProductPromotion :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $5
    AND active = TRUE
    )
INSERT INTO "product_promotion" (
  product_id,
  promotion_id,
  product_promotion_image,
  active
)
SELECT $1, $2, $3, $4 FROM t1
WHERE is_admin=1
RETURNING product_id, promotion_id, product_promotion_image, active
`

type AdminCreateProductPromotionParams struct {
	ProductID             int64       `json:"product_id"`
	PromotionID           int64       `json:"promotion_id"`
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                bool        `json:"active"`
	AdminID               int64       `json:"admin_id"`
}

func (q *Queries) AdminCreateProductPromotion(ctx context.Context, arg AdminCreateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, adminCreateProductPromotion,
		arg.ProductID,
		arg.PromotionID,
		arg.ProductPromotionImage,
		arg.Active,
		arg.AdminID,
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

const adminListProductPromotions = `-- name: AdminListProductPromotions :many
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $1
    AND active = TRUE
    )
SELECT 
pp.product_id, p.name AS product_name, 
pp.promotion_id, promo.name AS promotion_name,
pp.product_promotion_image, pp.active FROM "product_promotion" AS pp
LEFT JOIN "product" AS p ON p.id = pp.product_id
LEFT JOIN "promotion" AS promo ON promo.id = pp.promotion_id
WHERE (SELECT is_admin FROM t1) = 1
ORDER BY product_id
`

type AdminListProductPromotionsRow struct {
	ProductID             int64       `json:"product_id"`
	ProductName           null.String `json:"product_name"`
	PromotionID           int64       `json:"promotion_id"`
	PromotionName         null.String `json:"promotion_name"`
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                bool        `json:"active"`
}

func (q *Queries) AdminListProductPromotions(ctx context.Context, adminID int64) ([]AdminListProductPromotionsRow, error) {
	rows, err := q.db.Query(ctx, adminListProductPromotions, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AdminListProductPromotionsRow{}
	for rows.Next() {
		var i AdminListProductPromotionsRow
		if err := rows.Scan(
			&i.ProductID,
			&i.ProductName,
			&i.PromotionID,
			&i.PromotionName,
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

const adminUpdateProductPromotion = `-- name: AdminUpdateProductPromotion :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $5
    AND active = TRUE
    )
UPDATE "product_promotion"
SET
product_promotion_image = COALESCE($1,product_promotion_image),
active = COALESCE($2,active)
WHERE product_id = $3
AND promotion_id = $4
AND (SELECT is_admin FROM t1) = 1
RETURNING product_id, promotion_id, product_promotion_image, active
`

type AdminUpdateProductPromotionParams struct {
	ProductPromotionImage null.String `json:"product_promotion_image"`
	Active                null.Bool   `json:"active"`
	ProductID             int64       `json:"product_id"`
	PromotionID           int64       `json:"promotion_id"`
	AdminID               int64       `json:"admin_id"`
}

func (q *Queries) AdminUpdateProductPromotion(ctx context.Context, arg AdminUpdateProductPromotionParams) (ProductPromotion, error) {
	row := q.db.QueryRow(ctx, adminUpdateProductPromotion,
		arg.ProductPromotionImage,
		arg.Active,
		arg.ProductID,
		arg.PromotionID,
		arg.AdminID,
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
SELECT pp.product_id, promotion_id, product_promotion_image, pp.active, p.id, category_id, brand_id, p.name, p.description, p.active, p.created_at, p.updated_at, search, promo.id, promo.name, promo.description, discount_rate, promo.active, start_date, end_date, pi.id, pi.product_id, size_id, image_id, color_id, product_sku, qty_in_stock, price, pi.active, pi.created_at, pi.updated_at FROM "product_promotion" AS pp
LEFT JOIN "product" AS p ON p.id = pp.product_id
JOIN "promotion" AS promo ON promo.id = pp.promotion_id AND promo.active = TRUE AND promo.start_date <= CURRENT_DATE AND promo.end_date >= CURRENT_DATE
JOIN "product_item" AS pi ON pi.product_id = p.id AND pi.active =TRUE
WHERE pp.product_promotion_image IS NOT NULL AND pp.active = TRUE
ORDER BY promo.start_date DESC
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
	ID_2                  int64       `json:"id_2"`
	Name_2                string      `json:"name_2"`
	Description_2         string      `json:"description_2"`
	DiscountRate          int64       `json:"discount_rate"`
	Active_3              bool        `json:"active_3"`
	StartDate             time.Time   `json:"start_date"`
	EndDate               time.Time   `json:"end_date"`
	ID_3                  int64       `json:"id_3"`
	ProductID_2           int64       `json:"product_id_2"`
	SizeID                int64       `json:"size_id"`
	ImageID               int64       `json:"image_id"`
	ColorID               int64       `json:"color_id"`
	ProductSku            int64       `json:"product_sku"`
	QtyInStock            int32       `json:"qty_in_stock"`
	Price                 string      `json:"price"`
	Active_4              bool        `json:"active_4"`
	CreatedAt_2           time.Time   `json:"created_at_2"`
	UpdatedAt_2           time.Time   `json:"updated_at_2"`
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
			&i.ID_2,
			&i.Name_2,
			&i.Description_2,
			&i.DiscountRate,
			&i.Active_3,
			&i.StartDate,
			&i.EndDate,
			&i.ID_3,
			&i.ProductID_2,
			&i.SizeID,
			&i.ImageID,
			&i.ColorID,
			&i.ProductSku,
			&i.QtyInStock,
			&i.Price,
			&i.Active_4,
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
