// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: address.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const createAddress = `-- name: CreateAddress :one
INSERT INTO "address" (
  address_line,
  region,
  city
) VALUES (
  $1, $2, $3
)
RETURNING id, address_line, region, city, created_at, updated_at
`

type CreateAddressParams struct {
	AddressLine string `json:"address_line"`
	Region      string `json:"region"`
	City        string `json:"city"`
}

func (q *Queries) CreateAddress(ctx context.Context, arg CreateAddressParams) (Address, error) {
	row := q.db.QueryRow(ctx, createAddress, arg.AddressLine, arg.Region, arg.City)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.AddressLine,
		&i.Region,
		&i.City,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAddress = `-- name: DeleteAddress :exec
DELETE FROM "address"
WHERE id = $1
`

func (q *Queries) DeleteAddress(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteAddress, id)
	return err
}

const getAddress = `-- name: GetAddress :one
SELECT id, address_line, region, city, created_at, updated_at FROM "address"
WHERE id = $1 
LIMIT 1
`

func (q *Queries) GetAddress(ctx context.Context, id int64) (Address, error) {
	row := q.db.QueryRow(ctx, getAddress, id)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.AddressLine,
		&i.Region,
		&i.City,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAddressByCity = `-- name: GetAddressByCity :one
SELECT id, address_line, region, city, created_at, updated_at FROM "address"
WHERE city = $1 
LIMIT 1
`

func (q *Queries) GetAddressByCity(ctx context.Context, city string) (Address, error) {
	row := q.db.QueryRow(ctx, getAddressByCity, city)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.AddressLine,
		&i.Region,
		&i.City,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listAddressesByCity = `-- name: ListAddressesByCity :many
SELECT id, address_line, region, city, created_at, updated_at FROM "address"
WHERE city = $1
ORDER BY id
LIMIT $2
OFFSET $3
`

type ListAddressesByCityParams struct {
	City   string `json:"city"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) ListAddressesByCity(ctx context.Context, arg ListAddressesByCityParams) ([]Address, error) {
	rows, err := q.db.Query(ctx, listAddressesByCity, arg.City, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Address{}
	for rows.Next() {
		var i Address
		if err := rows.Scan(
			&i.ID,
			&i.AddressLine,
			&i.Region,
			&i.City,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const updateAddress = `-- name: UpdateAddress :one
UPDATE "address"
SET 
address_line = COALESCE($1,address_line),
region = COALESCE($2,region),
city = COALESCE($3,city)
WHERE id = $4
RETURNING id, address_line, region, city, created_at, updated_at
`

type UpdateAddressParams struct {
	AddressLine null.String `json:"address_line"`
	Region      null.String `json:"region"`
	City        null.String `json:"city"`
	ID          int64       `json:"id"`
}

func (q *Queries) UpdateAddress(ctx context.Context, arg UpdateAddressParams) (Address, error) {
	row := q.db.QueryRow(ctx, updateAddress,
		arg.AddressLine,
		arg.Region,
		arg.City,
		arg.ID,
	)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.AddressLine,
		&i.Region,
		&i.City,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
