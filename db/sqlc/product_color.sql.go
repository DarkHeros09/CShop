// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: product_color.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createProductColor = `-- name: CreateProductColor :one
INSERT INTO "product_color" (
  color_value
) VALUES (
  $1
)
RETURNING id, color_value
`

func (q *Queries) CreateProductColor(ctx context.Context, colorValue string) (ProductColor, error) {
	row := q.db.QueryRow(ctx, createProductColor, colorValue)
	var i ProductColor
	err := row.Scan(&i.ID, &i.ColorValue)
	return i, err
}

const deleteProductColor = `-- name: DeleteProductColor :exec
DELETE FROM "product_color"
WHERE id = $1
`

func (q *Queries) DeleteProductColor(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProductColor, id)
	return err
}

const getProductColor = `-- name: GetProductColor :one
SELECT id, color_value FROM "product_color"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductColor(ctx context.Context, id int64) (ProductColor, error) {
	row := q.db.QueryRow(ctx, getProductColor, id)
	var i ProductColor
	err := row.Scan(&i.ID, &i.ColorValue)
	return i, err
}

const updateProductColor = `-- name: UpdateProductColor :one
UPDATE "product_color"
SET 
color_value = COALESCE($1,color_value)
WHERE id = $2
RETURNING id, color_value
`

type UpdateProductColorParams struct {
	ColorValue null.String `json:"color_value"`
	ID         int64       `json:"id"`
}

func (q *Queries) UpdateProductColor(ctx context.Context, arg UpdateProductColorParams) (ProductColor, error) {
	row := q.db.QueryRow(ctx, updateProductColor, arg.ColorValue, arg.ID)
	var i ProductColor
	err := row.Scan(&i.ID, &i.ColorValue)
	return i, err
}
