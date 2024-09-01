// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: product_image.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const adminCreateProductImages = `-- name: AdminCreateProductImages :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $4
    AND active = TRUE
    )
INSERT INTO "product_image" (
  product_image_1,
  product_image_2,
  product_image_3
)
SELECT $1, $2, $3 FROM t1
WHERE is_admin=1
RETURNING id, product_image_1, product_image_2, product_image_3
`

type AdminCreateProductImagesParams struct {
	ProductImage1 string `json:"product_image_1"`
	ProductImage2 string `json:"product_image_2"`
	ProductImage3 string `json:"product_image_3"`
	AdminID       int64  `json:"admin_id"`
}

func (q *Queries) AdminCreateProductImages(ctx context.Context, arg AdminCreateProductImagesParams) (ProductImage, error) {
	row := q.db.QueryRow(ctx, adminCreateProductImages,
		arg.ProductImage1,
		arg.ProductImage2,
		arg.ProductImage3,
		arg.AdminID,
	)
	var i ProductImage
	err := row.Scan(
		&i.ID,
		&i.ProductImage1,
		&i.ProductImage2,
		&i.ProductImage3,
	)
	return i, err
}

const adminUpdateProductImage = `-- name: AdminUpdateProductImage :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $5
    AND active = TRUE
    )
UPDATE "product_image"
SET 
product_image_1 = COALESCE($1,product_image_1),
product_image_2 = COALESCE($2,product_image_2),
product_image_3 = COALESCE($3,product_image_3)
WHERE "product_image".id = $4
AND (SELECT is_admin FROM t1) = 1
RETURNING id, product_image_1, product_image_2, product_image_3
`

type AdminUpdateProductImageParams struct {
	ProductImage1 null.String `json:"product_image_1"`
	ProductImage2 null.String `json:"product_image_2"`
	ProductImage3 null.String `json:"product_image_3"`
	ID            int64       `json:"id"`
	AdminID       int64       `json:"admin_id"`
}

func (q *Queries) AdminUpdateProductImage(ctx context.Context, arg AdminUpdateProductImageParams) (ProductImage, error) {
	row := q.db.QueryRow(ctx, adminUpdateProductImage,
		arg.ProductImage1,
		arg.ProductImage2,
		arg.ProductImage3,
		arg.ID,
		arg.AdminID,
	)
	var i ProductImage
	err := row.Scan(
		&i.ID,
		&i.ProductImage1,
		&i.ProductImage2,
		&i.ProductImage3,
	)
	return i, err
}

const createProductImage = `-- name: CreateProductImage :one
INSERT INTO "product_image" (
  product_image_1,
  product_image_2,
  product_image_3
) VALUES (
  $1, $2, $3
)
RETURNING id, product_image_1, product_image_2, product_image_3
`

type CreateProductImageParams struct {
	ProductImage1 string `json:"product_image_1"`
	ProductImage2 string `json:"product_image_2"`
	ProductImage3 string `json:"product_image_3"`
}

func (q *Queries) CreateProductImage(ctx context.Context, arg CreateProductImageParams) (ProductImage, error) {
	row := q.db.QueryRow(ctx, createProductImage, arg.ProductImage1, arg.ProductImage2, arg.ProductImage3)
	var i ProductImage
	err := row.Scan(
		&i.ID,
		&i.ProductImage1,
		&i.ProductImage2,
		&i.ProductImage3,
	)
	return i, err
}

const deleteProductImage = `-- name: DeleteProductImage :exec
DELETE FROM "product_image"
WHERE id = $1
`

func (q *Queries) DeleteProductImage(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProductImage, id)
	return err
}

const getProductImage = `-- name: GetProductImage :one
SELECT id, product_image_1, product_image_2, product_image_3 FROM "product_image"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductImage(ctx context.Context, id int64) (ProductImage, error) {
	row := q.db.QueryRow(ctx, getProductImage, id)
	var i ProductImage
	err := row.Scan(
		&i.ID,
		&i.ProductImage1,
		&i.ProductImage2,
		&i.ProductImage3,
	)
	return i, err
}

const listProductImagesNextPage = `-- name: ListProductImagesNextPage :many
WITH t1 AS(
SELECT 
 pimg.id, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM "product_image" AS pimg
WHERE
 pimg.id < $2 
ORDER BY id DESC
LIMIT $1 +1
)

SELECT id, product_image_1, product_image_2, product_image_3,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type ListProductImagesNextPageParams struct {
	Limit int32 `json:"limit"`
	ID    int64 `json:"id"`
}

type ListProductImagesNextPageRow struct {
	ID            int64  `json:"id"`
	ProductImage1 string `json:"product_image_1"`
	ProductImage2 string `json:"product_image_2"`
	ProductImage3 string `json:"product_image_3"`
	NextAvailable bool   `json:"next_available"`
}

func (q *Queries) ListProductImagesNextPage(ctx context.Context, arg ListProductImagesNextPageParams) ([]ListProductImagesNextPageRow, error) {
	rows, err := q.db.Query(ctx, listProductImagesNextPage, arg.Limit, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductImagesNextPageRow{}
	for rows.Next() {
		var i ListProductImagesNextPageRow
		if err := rows.Scan(
			&i.ID,
			&i.ProductImage1,
			&i.ProductImage2,
			&i.ProductImage3,
			&i.NextAvailable,
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

const listProductImagesV2 = `-- name: ListProductImagesV2 :many
WITH t1 AS(
SELECT 
 pimg.id, pimg.product_image_1, pimg.product_image_2, pimg.product_image_3
FROM "product_image" AS pimg
ORDER BY id DESC
LIMIT $1 +1
)

SELECT id, product_image_1, product_image_2, product_image_3,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type ListProductImagesV2Row struct {
	ID            int64  `json:"id"`
	ProductImage1 string `json:"product_image_1"`
	ProductImage2 string `json:"product_image_2"`
	ProductImage3 string `json:"product_image_3"`
	NextAvailable bool   `json:"next_available"`
}

func (q *Queries) ListProductImagesV2(ctx context.Context, limit int32) ([]ListProductImagesV2Row, error) {
	rows, err := q.db.Query(ctx, listProductImagesV2, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductImagesV2Row{}
	for rows.Next() {
		var i ListProductImagesV2Row
		if err := rows.Scan(
			&i.ID,
			&i.ProductImage1,
			&i.ProductImage2,
			&i.ProductImage3,
			&i.NextAvailable,
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

const updateProductImage = `-- name: UpdateProductImage :one
UPDATE "product_image"
SET 
product_image_1 = COALESCE($1,product_image_1),
product_image_2 = COALESCE($2,product_image_2),
product_image_3 = COALESCE($3,product_image_3)
WHERE id = $4
RETURNING id, product_image_1, product_image_2, product_image_3
`

type UpdateProductImageParams struct {
	ProductImage1 null.String `json:"product_image_1"`
	ProductImage2 null.String `json:"product_image_2"`
	ProductImage3 null.String `json:"product_image_3"`
	ID            int64       `json:"id"`
}

func (q *Queries) UpdateProductImage(ctx context.Context, arg UpdateProductImageParams) (ProductImage, error) {
	row := q.db.QueryRow(ctx, updateProductImage,
		arg.ProductImage1,
		arg.ProductImage2,
		arg.ProductImage3,
		arg.ID,
	)
	var i ProductImage
	err := row.Scan(
		&i.ID,
		&i.ProductImage1,
		&i.ProductImage2,
		&i.ProductImage3,
	)
	return i, err
}
