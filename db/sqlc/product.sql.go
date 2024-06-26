// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: product.sql

package db

import (
	"context"
	"time"

	null "github.com/guregu/null/v5"
)

const adminCreateProduct = `-- name: AdminCreateProduct :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $6
    AND active = TRUE
    )
INSERT INTO "product" (
  category_id,
  brand_id,
  name,
  description,
  active
)
SELECT $1, $2, $3, $4, $5 FROM t1
WHERE is_admin=1
RETURNING id, category_id, brand_id, name, description, active, created_at, updated_at, search
`

type AdminCreateProductParams struct {
	CategoryID  int64  `json:"category_id"`
	BrandID     int64  `json:"brand_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	AdminID     int64  `json:"admin_id"`
}

func (q *Queries) AdminCreateProduct(ctx context.Context, arg AdminCreateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, adminCreateProduct,
		arg.CategoryID,
		arg.BrandID,
		arg.Name,
		arg.Description,
		arg.Active,
		arg.AdminID,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.CategoryID,
		&i.BrandID,
		&i.Name,
		&i.Description,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Search,
	)
	return i, err
}

const createProduct = `-- name: CreateProduct :one
INSERT INTO "product" (
  category_id,
  brand_id,
  name,
  description,
  -- product_image,
  active
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, category_id, brand_id, name, description, active, created_at, updated_at, search
`

type CreateProductParams struct {
	CategoryID  int64  `json:"category_id"`
	BrandID     int64  `json:"brand_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, createProduct,
		arg.CategoryID,
		arg.BrandID,
		arg.Name,
		arg.Description,
		arg.Active,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.CategoryID,
		&i.BrandID,
		&i.Name,
		&i.Description,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Search,
	)
	return i, err
}

const deleteProduct = `-- name: DeleteProduct :exec
DELETE FROM "product"
WHERE id = $1
`

func (q *Queries) DeleteProduct(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProduct, id)
	return err
}

const getProduct = `-- name: GetProduct :one
SELECT id, category_id, brand_id, name, description, active, created_at, updated_at, search FROM "product"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProduct(ctx context.Context, id int64) (Product, error) {
	row := q.db.QueryRow(ctx, getProduct, id)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.CategoryID,
		&i.BrandID,
		&i.Name,
		&i.Description,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Search,
	)
	return i, err
}

const getProductsByIDs = `-- name: GetProductsByIDs :many
SELECT id, category_id, brand_id, name, description, active, created_at, updated_at, search FROM "product"
WHERE id = ANY($1::bigint[])
`

func (q *Queries) GetProductsByIDs(ctx context.Context, ids []int64) ([]Product, error) {
	rows, err := q.db.Query(ctx, getProductsByIDs, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.CategoryID,
			&i.BrandID,
			&i.Name,
			&i.Description,
			&i.Active,
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

const listProducts = `-- name: ListProducts :many
SELECT id, category_id, brand_id, name, description, active, created_at, updated_at, search ,
COUNT(*) OVER() AS total_count
FROM "product"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListProductsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListProductsRow struct {
	ID          int64       `json:"id"`
	CategoryID  int64       `json:"category_id"`
	BrandID     int64       `json:"brand_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Active      bool        `json:"active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Search      null.String `json:"search"`
	TotalCount  int64       `json:"total_count"`
}

// WITH total_records AS (
//
//	SELECT COUNT(id)
//	FROM "product"
//
// ),
// list_products AS (
func (q *Queries) ListProducts(ctx context.Context, arg ListProductsParams) ([]ListProductsRow, error) {
	rows, err := q.db.Query(ctx, listProducts, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductsRow{}
	for rows.Next() {
		var i ListProductsRow
		if err := rows.Scan(
			&i.ID,
			&i.CategoryID,
			&i.BrandID,
			&i.Name,
			&i.Description,
			&i.Active,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Search,
			&i.TotalCount,
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

const updateProduct = `-- name: UpdateProduct :one

UPDATE "product"
SET
category_id = COALESCE($1,category_id),
brand_id = COALESCE($2,brand_id),
name = COALESCE($3,name),
description = COALESCE($4,description),
active = COALESCE($5,active),
updated_at = now()
WHERE id = $6
RETURNING id, category_id, brand_id, name, description, active, created_at, updated_at, search
`

type UpdateProductParams struct {
	CategoryID  null.Int    `json:"category_id"`
	BrandID     null.Int    `json:"brand_id"`
	Name        null.String `json:"name"`
	Description null.String `json:"description"`
	Active      null.Bool   `json:"active"`
	ID          int64       `json:"id"`
}

// )
// SELECT *
// FROM list_products, total_records;
// product_image = COALESCE(sqlc.narg(product_image),product_image),
func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, updateProduct,
		arg.CategoryID,
		arg.BrandID,
		arg.Name,
		arg.Description,
		arg.Active,
		arg.ID,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.CategoryID,
		&i.BrandID,
		&i.Name,
		&i.Description,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Search,
	)
	return i, err
}
