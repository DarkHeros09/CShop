// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: product_configuration.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createProductConfiguration = `-- name: CreateProductConfiguration :one
INSERT INTO "product_configuration" (
  product_item_id,
  variation_option_id
) VALUES (
  $1, $2
)
RETURNING product_item_id, variation_option_id
`

type CreateProductConfigurationParams struct {
	ProductItemID     int64 `json:"product_item_id"`
	VariationOptionID int64 `json:"variation_option_id"`
}

func (q *Queries) CreateProductConfiguration(ctx context.Context, arg CreateProductConfigurationParams) (ProductConfiguration, error) {
	row := q.db.QueryRow(ctx, createProductConfiguration, arg.ProductItemID, arg.VariationOptionID)
	var i ProductConfiguration
	err := row.Scan(&i.ProductItemID, &i.VariationOptionID)
	return i, err
}

const deleteProductConfiguration = `-- name: DeleteProductConfiguration :exec
DELETE FROM "product_configuration"
WHERE product_item_id = $1
AND variation_option_id = $2
`

type DeleteProductConfigurationParams struct {
	ProductItemID     int64 `json:"product_item_id"`
	VariationOptionID int64 `json:"variation_option_id"`
}

func (q *Queries) DeleteProductConfiguration(ctx context.Context, arg DeleteProductConfigurationParams) error {
	_, err := q.db.Exec(ctx, deleteProductConfiguration, arg.ProductItemID, arg.VariationOptionID)
	return err
}

const getProductConfiguration = `-- name: GetProductConfiguration :one
SELECT product_item_id, variation_option_id FROM "product_configuration"
WHERE product_item_id = $1 
AND variation_option_id = $2
LIMIT 1
`

type GetProductConfigurationParams struct {
	ProductItemID     int64 `json:"product_item_id"`
	VariationOptionID int64 `json:"variation_option_id"`
}

func (q *Queries) GetProductConfiguration(ctx context.Context, arg GetProductConfigurationParams) (ProductConfiguration, error) {
	row := q.db.QueryRow(ctx, getProductConfiguration, arg.ProductItemID, arg.VariationOptionID)
	var i ProductConfiguration
	err := row.Scan(&i.ProductItemID, &i.VariationOptionID)
	return i, err
}

const listProductConfigurations = `-- name: ListProductConfigurations :many
SELECT product_item_id, variation_option_id FROM "product_configuration"
ORDER BY product_item_id
LIMIT $1
OFFSET $2
`

type ListProductConfigurationsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListProductConfigurations(ctx context.Context, arg ListProductConfigurationsParams) ([]ProductConfiguration, error) {
	rows, err := q.db.Query(ctx, listProductConfigurations, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProductConfiguration{}
	for rows.Next() {
		var i ProductConfiguration
		if err := rows.Scan(&i.ProductItemID, &i.VariationOptionID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateProductConfiguration = `-- name: UpdateProductConfiguration :one
UPDATE "product_configuration"
SET
variation_option_id = COALESCE($1,variation_option_id)
WHERE product_item_id = $2
RETURNING product_item_id, variation_option_id
`

type UpdateProductConfigurationParams struct {
	VariationOptionID null.Int `json:"variation_option_id"`
	ProductItemID     int64    `json:"product_item_id"`
}

func (q *Queries) UpdateProductConfiguration(ctx context.Context, arg UpdateProductConfigurationParams) (ProductConfiguration, error) {
	row := q.db.QueryRow(ctx, updateProductConfiguration, arg.VariationOptionID, arg.ProductItemID)
	var i ProductConfiguration
	err := row.Scan(&i.ProductItemID, &i.VariationOptionID)
	return i, err
}
