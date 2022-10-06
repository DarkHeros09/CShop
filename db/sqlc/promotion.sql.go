// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: promotion.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const createPromotion = `-- name: CreatePromotion :one
INSERT INTO "promotion" (
  name,
  description,
  discount_rate,
  active,
  start_date,
  end_date
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, name, description, discount_rate, active, start_date, end_date
`

type CreatePromotionParams struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DiscountRate int32     `json:"discount_rate"`
	Active       bool      `json:"active"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
}

func (q *Queries) CreatePromotion(ctx context.Context, arg CreatePromotionParams) (Promotion, error) {
	row := q.db.QueryRowContext(ctx, createPromotion,
		arg.Name,
		arg.Description,
		arg.DiscountRate,
		arg.Active,
		arg.StartDate,
		arg.EndDate,
	)
	var i Promotion
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.DiscountRate,
		&i.Active,
		&i.StartDate,
		&i.EndDate,
	)
	return i, err
}

const deletePromotion = `-- name: DeletePromotion :exec
DELETE FROM "promotion"
WHERE id = $1
`

func (q *Queries) DeletePromotion(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deletePromotion, id)
	return err
}

const getPromotion = `-- name: GetPromotion :one
SELECT id, name, description, discount_rate, active, start_date, end_date FROM "promotion"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetPromotion(ctx context.Context, id int64) (Promotion, error) {
	row := q.db.QueryRowContext(ctx, getPromotion, id)
	var i Promotion
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.DiscountRate,
		&i.Active,
		&i.StartDate,
		&i.EndDate,
	)
	return i, err
}

const listPromotions = `-- name: ListPromotions :many
SELECT id, name, description, discount_rate, active, start_date, end_date FROM "promotion"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListPromotionsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListPromotions(ctx context.Context, arg ListPromotionsParams) ([]Promotion, error) {
	rows, err := q.db.QueryContext(ctx, listPromotions, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Promotion{}
	for rows.Next() {
		var i Promotion
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.DiscountRate,
			&i.Active,
			&i.StartDate,
			&i.EndDate,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePromotion = `-- name: UpdatePromotion :one
UPDATE "promotion"
SET
name = COALESCE($1,name),
description = COALESCE($2,description),
discount_rate = COALESCE($3,discount_rate),
active = COALESCE($4,active),
start_date = COALESCE($5,start_date),
end_date = COALESCE($6,end_date)
WHERE id = $7
RETURNING id, name, description, discount_rate, active, start_date, end_date
`

type UpdatePromotionParams struct {
	Name         sql.NullString `json:"name"`
	Description  sql.NullString `json:"description"`
	DiscountRate sql.NullInt32  `json:"discount_rate"`
	Active       sql.NullBool   `json:"active"`
	StartDate    sql.NullTime   `json:"start_date"`
	EndDate      sql.NullTime   `json:"end_date"`
	ID           int64          `json:"id"`
}

func (q *Queries) UpdatePromotion(ctx context.Context, arg UpdatePromotionParams) (Promotion, error) {
	row := q.db.QueryRowContext(ctx, updatePromotion,
		arg.Name,
		arg.Description,
		arg.DiscountRate,
		arg.Active,
		arg.StartDate,
		arg.EndDate,
		arg.ID,
	)
	var i Promotion
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.DiscountRate,
		&i.Active,
		&i.StartDate,
		&i.EndDate,
	)
	return i, err
}
