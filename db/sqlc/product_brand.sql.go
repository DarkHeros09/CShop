// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: product_brand.sql

package db

import (
	"context"
)

const createProductBrand = `-- name: CreateProductBrand :one
INSERT INTO "product_brand" (
  brand_name,
  brand_image
) VALUES (
  $1, $2
)
ON CONFLICT(brand_name) DO UPDATE SET 
brand_name = EXCLUDED.brand_name,
brand_image = EXCLUDED.brand_image
RETURNING id, brand_name, brand_image
`

type CreateProductBrandParams struct {
	BrandName  string `json:"brand_name"`
	BrandImage string `json:"brand_image"`
}

func (q *Queries) CreateProductBrand(ctx context.Context, arg CreateProductBrandParams) (ProductBrand, error) {
	row := q.db.QueryRow(ctx, createProductBrand, arg.BrandName, arg.BrandImage)
	var i ProductBrand
	err := row.Scan(&i.ID, &i.BrandName, &i.BrandImage)
	return i, err
}

const deleteProductBrand = `-- name: DeleteProductBrand :exec
DELETE FROM "product_brand"
WHERE id = $1
`

func (q *Queries) DeleteProductBrand(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProductBrand, id)
	return err
}

const getProductBrand = `-- name: GetProductBrand :one
SELECT id, brand_name, brand_image FROM "product_brand"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductBrand(ctx context.Context, id int64) (ProductBrand, error) {
	row := q.db.QueryRow(ctx, getProductBrand, id)
	var i ProductBrand
	err := row.Scan(&i.ID, &i.BrandName, &i.BrandImage)
	return i, err
}

const listProductBrands = `-- name: ListProductBrands :many
SELECT id, brand_name, brand_image FROM "product_brand"
ORDER BY id
`

func (q *Queries) ListProductBrands(ctx context.Context) ([]ProductBrand, error) {
	rows, err := q.db.Query(ctx, listProductBrands)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProductBrand{}
	for rows.Next() {
		var i ProductBrand
		if err := rows.Scan(&i.ID, &i.BrandName, &i.BrandImage); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateProductBrand = `-- name: UpdateProductBrand :one

UPDATE "product_brand"
SET brand_name = $1
WHERE id = $2
RETURNING id, brand_name, brand_image
`

type UpdateProductBrandParams struct {
	BrandName string `json:"brand_name"`
	ID        int64  `json:"id"`
}

// LIMIT $1
// OFFSET $2;
func (q *Queries) UpdateProductBrand(ctx context.Context, arg UpdateProductBrandParams) (ProductBrand, error) {
	row := q.db.QueryRow(ctx, updateProductBrand, arg.BrandName, arg.ID)
	var i ProductBrand
	err := row.Scan(&i.ID, &i.BrandName, &i.BrandImage)
	return i, err
}
