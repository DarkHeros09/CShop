// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
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

const adminDeleteProduct = `-- name: AdminDeleteProduct :exec
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $2
    AND active = TRUE
    )
DELETE FROM "product" AS p
WHERE p.id = $1
AND (SELECT is_admin FROM t1) = 1
`

type AdminDeleteProductParams struct {
	ID      int64 `json:"id"`
	AdminID int64 `json:"admin_id"`
}

func (q *Queries) AdminDeleteProduct(ctx context.Context, arg AdminDeleteProductParams) error {
	_, err := q.db.Exec(ctx, adminDeleteProduct, arg.ID, arg.AdminID)
	return err
}

const adminUpdateProduct = `-- name: AdminUpdateProduct :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $7
    AND active = TRUE
    )
UPDATE "product"
SET
category_id = COALESCE($1,category_id),
brand_id = COALESCE($2,brand_id),
name = COALESCE($3,name),
description = COALESCE($4,description),
active = COALESCE($5,active),
updated_at = now()
WHERE "product".id = $6
AND (SELECT is_admin FROM t1) = 1
RETURNING id, category_id, brand_id, name, description, active, created_at, updated_at, search
`

type AdminUpdateProductParams struct {
	CategoryID  null.Int    `json:"category_id"`
	BrandID     null.Int    `json:"brand_id"`
	Name        null.String `json:"name"`
	Description null.String `json:"description"`
	Active      null.Bool   `json:"active"`
	ID          int64       `json:"id"`
	AdminID     int64       `json:"admin_id"`
}

// product_image = COALESCE(sqlc.narg(product_image),product_image),
func (q *Queries) AdminUpdateProduct(ctx context.Context, arg AdminUpdateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, adminUpdateProduct,
		arg.CategoryID,
		arg.BrandID,
		arg.Name,
		arg.Description,
		arg.Active,
		arg.ID,
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

const listProductsNextPage = `-- name: ListProductsNextPage :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
WHERE
 p.id < $2 
ORDER BY id DESC
LIMIT $1 +1
)

SELECT id, name, description, category_id, brand_id, active, created_at, updated_at,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type ListProductsNextPageParams struct {
	Limit int32 `json:"limit"`
	ID    int64 `json:"id"`
}

type ListProductsNextPageRow struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CategoryID    int64     `json:"category_id"`
	BrandID       int64     `json:"brand_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	NextAvailable bool      `json:"next_available"`
}

func (q *Queries) ListProductsNextPage(ctx context.Context, arg ListProductsNextPageParams) ([]ListProductsNextPageRow, error) {
	rows, err := q.db.Query(ctx, listProductsNextPage, arg.Limit, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductsNextPageRow{}
	for rows.Next() {
		var i ListProductsNextPageRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.CategoryID,
			&i.BrandID,
			&i.Active,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const listProductsV2 = `-- name: ListProductsV2 :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
ORDER BY id DESC
LIMIT $1 +1
)

SELECT id, name, description, category_id, brand_id, active, created_at, updated_at,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type ListProductsV2Row struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CategoryID    int64     `json:"category_id"`
	BrandID       int64     `json:"brand_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	NextAvailable bool      `json:"next_available"`
}

func (q *Queries) ListProductsV2(ctx context.Context, limit int32) ([]ListProductsV2Row, error) {
	rows, err := q.db.Query(ctx, listProductsV2, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProductsV2Row{}
	for rows.Next() {
		var i ListProductsV2Row
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.CategoryID,
			&i.BrandID,
			&i.Active,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const searchProducts = `-- name: SearchProducts :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
WHERE 
p.search @@ 
CASE
    WHEN char_length($2::VARCHAR) > 0 THEN to_tsquery(concat($2, ':*'))
    ELSE to_tsquery($2)
END
ORDER BY 
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length($2) > 0 THEN to_tsquery(concat($2, ':*'))
    ELSE to_tsquery($2)
END
) DESC
LIMIT $1 +1
)

SELECT id, name, description, category_id, brand_id, active, created_at, updated_at,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type SearchProductsParams struct {
	Limit int32  `json:"limit"`
	Query string `json:"query"`
}

type SearchProductsRow struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CategoryID    int64     `json:"category_id"`
	BrandID       int64     `json:"brand_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	NextAvailable bool      `json:"next_available"`
}

func (q *Queries) SearchProducts(ctx context.Context, arg SearchProductsParams) ([]SearchProductsRow, error) {
	rows, err := q.db.Query(ctx, searchProducts, arg.Limit, arg.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SearchProductsRow{}
	for rows.Next() {
		var i SearchProductsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.CategoryID,
			&i.BrandID,
			&i.Active,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const searchProductsNextPage = `-- name: SearchProductsNextPage :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p

WHERE 
p.id < $2 AND
p.search @@ 
CASE
    WHEN char_length($3::VARCHAR) > 0 THEN to_tsquery(concat($3, ':*')::VARCHAR)
    ELSE to_tsquery($3::VARCHAR)
END
ORDER BY 
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length($3::VARCHAR) > 0 THEN to_tsquery(concat($3, ':*')::VARCHAR)
    ELSE to_tsquery($3::VARCHAR)
END
) DESC
LIMIT $1 +1
)

SELECT id, name, description, category_id, brand_id, active, created_at, updated_at,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1
`

type SearchProductsNextPageParams struct {
	Limit     int32  `json:"limit"`
	ProductID int64  `json:"product_id"`
	Query     string `json:"query"`
}

type SearchProductsNextPageRow struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CategoryID    int64     `json:"category_id"`
	BrandID       int64     `json:"brand_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	NextAvailable bool      `json:"next_available"`
}

func (q *Queries) SearchProductsNextPage(ctx context.Context, arg SearchProductsNextPageParams) ([]SearchProductsNextPageRow, error) {
	rows, err := q.db.Query(ctx, searchProductsNextPage, arg.Limit, arg.ProductID, arg.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SearchProductsNextPageRow{}
	for rows.Next() {
		var i SearchProductsNextPageRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.CategoryID,
			&i.BrandID,
			&i.Active,
			&i.CreatedAt,
			&i.UpdatedAt,
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
