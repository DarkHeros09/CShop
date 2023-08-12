// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: product_category.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createProductCategory = `-- name: CreateProductCategory :one
INSERT INTO "product_category" (
  parent_category_id,
  category_name,
  category_image
) VALUES (
  $1, $2, $3
)
ON CONFLICT(category_name) DO UPDATE SET 
category_name = EXCLUDED.category_name,
category_image = EXCLUDED.category_image
RETURNING id, parent_category_id, category_name, category_image
`

type CreateProductCategoryParams struct {
	ParentCategoryID null.Int `json:"parent_category_id"`
	CategoryName     string   `json:"category_name"`
	CategoryImage    string   `json:"category_image"`
}

func (q *Queries) CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error) {
	row := q.db.QueryRow(ctx, createProductCategory, arg.ParentCategoryID, arg.CategoryName, arg.CategoryImage)
	var i ProductCategory
	err := row.Scan(
		&i.ID,
		&i.ParentCategoryID,
		&i.CategoryName,
		&i.CategoryImage,
	)
	return i, err
}

const deleteProductCategory = `-- name: DeleteProductCategory :exec
DELETE FROM "product_category"
WHERE id = $1
AND ( parent_category_id is NULL OR parent_category_id = $2 )
`

type DeleteProductCategoryParams struct {
	ID               int64    `json:"id"`
	ParentCategoryID null.Int `json:"parent_category_id"`
}

func (q *Queries) DeleteProductCategory(ctx context.Context, arg DeleteProductCategoryParams) error {
	_, err := q.db.Exec(ctx, deleteProductCategory, arg.ID, arg.ParentCategoryID)
	return err
}

const getProductCategory = `-- name: GetProductCategory :one
SELECT id, parent_category_id, category_name, category_image FROM "product_category"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductCategory(ctx context.Context, id int64) (ProductCategory, error) {
	row := q.db.QueryRow(ctx, getProductCategory, id)
	var i ProductCategory
	err := row.Scan(
		&i.ID,
		&i.ParentCategoryID,
		&i.CategoryName,
		&i.CategoryImage,
	)
	return i, err
}

const getProductCategoryByParent = `-- name: GetProductCategoryByParent :one
SELECT id, parent_category_id, category_name, category_image FROM "product_category"
WHERE id = $1
And parent_category_id = $2
LIMIT 1
`

type GetProductCategoryByParentParams struct {
	ID               int64    `json:"id"`
	ParentCategoryID null.Int `json:"parent_category_id"`
}

func (q *Queries) GetProductCategoryByParent(ctx context.Context, arg GetProductCategoryByParentParams) (ProductCategory, error) {
	row := q.db.QueryRow(ctx, getProductCategoryByParent, arg.ID, arg.ParentCategoryID)
	var i ProductCategory
	err := row.Scan(
		&i.ID,
		&i.ParentCategoryID,
		&i.CategoryName,
		&i.CategoryImage,
	)
	return i, err
}

const listProductCategories = `-- name: ListProductCategories :many
SELECT id, parent_category_id, category_name, category_image FROM "product_category"
ORDER BY id
`

func (q *Queries) ListProductCategories(ctx context.Context) ([]ProductCategory, error) {
	rows, err := q.db.Query(ctx, listProductCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProductCategory{}
	for rows.Next() {
		var i ProductCategory
		if err := rows.Scan(
			&i.ID,
			&i.ParentCategoryID,
			&i.CategoryName,
			&i.CategoryImage,
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

const listProductCategoriesByParent = `-- name: ListProductCategoriesByParent :many

SELECT id, parent_category_id, category_name, category_image FROM "product_category"
WHERE parent_category_id = $1
ORDER BY id
`

// LIMIT $1
// OFFSET $2;
func (q *Queries) ListProductCategoriesByParent(ctx context.Context, parentCategoryID null.Int) ([]ProductCategory, error) {
	rows, err := q.db.Query(ctx, listProductCategoriesByParent, parentCategoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProductCategory{}
	for rows.Next() {
		var i ProductCategory
		if err := rows.Scan(
			&i.ID,
			&i.ParentCategoryID,
			&i.CategoryName,
			&i.CategoryImage,
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

const updateProductCategory = `-- name: UpdateProductCategory :one

UPDATE "product_category"
SET category_name = $1
WHERE id = $2
AND
( parent_category_id is NULL OR parent_category_id = $3 )
RETURNING id, parent_category_id, category_name, category_image
`

type UpdateProductCategoryParams struct {
	CategoryName     string   `json:"category_name"`
	ID               int64    `json:"id"`
	ParentCategoryID null.Int `json:"parent_category_id"`
}

// LIMIT $2
// OFFSET $3;
func (q *Queries) UpdateProductCategory(ctx context.Context, arg UpdateProductCategoryParams) (ProductCategory, error) {
	row := q.db.QueryRow(ctx, updateProductCategory, arg.CategoryName, arg.ID, arg.ParentCategoryID)
	var i ProductCategory
	err := row.Scan(
		&i.ID,
		&i.ParentCategoryID,
		&i.CategoryName,
		&i.CategoryImage,
	)
	return i, err
}
