// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: product_size.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const createProductSize = `-- name: CreateProductSize :one
INSERT INTO "product_size" (
  size_value
) VALUES (
  $1
)
RETURNING id, size_value
`

func (q *Queries) CreateProductSize(ctx context.Context, sizeValue string) (ProductSize, error) {
	row := q.db.QueryRow(ctx, createProductSize, sizeValue)
	var i ProductSize
	err := row.Scan(&i.ID, &i.SizeValue)
	return i, err
}

const deleteProductSize = `-- name: DeleteProductSize :exec
DELETE FROM "product_size"
WHERE id = $1
`

func (q *Queries) DeleteProductSize(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProductSize, id)
	return err
}

const getProductSize = `-- name: GetProductSize :one
SELECT id, size_value FROM "product_size"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductSize(ctx context.Context, id int64) (ProductSize, error) {
	row := q.db.QueryRow(ctx, getProductSize, id)
	var i ProductSize
	err := row.Scan(&i.ID, &i.SizeValue)
	return i, err
}

const updateProductSize = `-- name: UpdateProductSize :one
UPDATE "product_size"
SET 
size_value = COALESCE($1,size_value)
WHERE id = $2
RETURNING id, size_value
`

type UpdateProductSizeParams struct {
	SizeValue null.String `json:"size_value"`
	ID        int64       `json:"id"`
}

func (q *Queries) UpdateProductSize(ctx context.Context, arg UpdateProductSizeParams) (ProductSize, error) {
	row := q.db.QueryRow(ctx, updateProductSize, arg.SizeValue, arg.ID)
	var i ProductSize
	err := row.Scan(&i.ID, &i.SizeValue)
	return i, err
}
