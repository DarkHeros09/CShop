// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: variation_option.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createVariationOption = `-- name: CreateVariationOption :one
INSERT INTO "variation_option" (
  variation_id,
  value
) VALUES (
  $1, $2
)
RETURNING id, variation_id, value
`

type CreateVariationOptionParams struct {
	VariationID int64  `json:"variation_id"`
	Value       string `json:"value"`
}

func (q *Queries) CreateVariationOption(ctx context.Context, arg CreateVariationOptionParams) (VariationOption, error) {
	row := q.db.QueryRow(ctx, createVariationOption, arg.VariationID, arg.Value)
	var i VariationOption
	err := row.Scan(&i.ID, &i.VariationID, &i.Value)
	return i, err
}

const deleteVariationOption = `-- name: DeleteVariationOption :exec
DELETE FROM "variation_option"
WHERE id = $1
`

func (q *Queries) DeleteVariationOption(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteVariationOption, id)
	return err
}

const getVariationOption = `-- name: GetVariationOption :one
SELECT id, variation_id, value FROM "variation_option"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetVariationOption(ctx context.Context, id int64) (VariationOption, error) {
	row := q.db.QueryRow(ctx, getVariationOption, id)
	var i VariationOption
	err := row.Scan(&i.ID, &i.VariationID, &i.Value)
	return i, err
}

const listVariationOptions = `-- name: ListVariationOptions :many
SELECT id, variation_id, value FROM "variation_option"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListVariationOptionsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListVariationOptions(ctx context.Context, arg ListVariationOptionsParams) ([]VariationOption, error) {
	rows, err := q.db.Query(ctx, listVariationOptions, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []VariationOption{}
	for rows.Next() {
		var i VariationOption
		if err := rows.Scan(&i.ID, &i.VariationID, &i.Value); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateVariationOption = `-- name: UpdateVariationOption :one
UPDATE "variation_option"
SET
variation_id = COALESCE($1,variation_id),
value = COALESCE($2,value)
WHERE id = $3
RETURNING id, variation_id, value
`

type UpdateVariationOptionParams struct {
	VariationID null.Int    `json:"variation_id"`
	Value       null.String `json:"value"`
	ID          int64       `json:"id"`
}

func (q *Queries) UpdateVariationOption(ctx context.Context, arg UpdateVariationOptionParams) (VariationOption, error) {
	row := q.db.QueryRow(ctx, updateVariationOption, arg.VariationID, arg.Value, arg.ID)
	var i VariationOption
	err := row.Scan(&i.ID, &i.VariationID, &i.Value)
	return i, err
}
