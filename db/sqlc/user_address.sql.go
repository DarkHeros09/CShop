// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: user_address.sql

package db

import (
	"context"

	"github.com/guregu/null"
)

const checkUserAddressDefaultAddress = `-- name: CheckUserAddressDefaultAddress :one
SELECT COUNT(*) FROM "user_address"
WHERE user_id = $1
`

func (q *Queries) CheckUserAddressDefaultAddress(ctx context.Context, userID int64) (int64, error) {
	row := q.db.QueryRow(ctx, checkUserAddressDefaultAddress, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createUserAddress = `-- name: CreateUserAddress :one
INSERT INTO "user_address" (
  user_id,
  address_id,
  default_address
) VALUES (
  $1, $2, $3
)
RETURNING user_id, address_id, default_address, created_at, updated_at
`

type CreateUserAddressParams struct {
	UserID         int64    `json:"user_id"`
	AddressID      int64    `json:"address_id"`
	DefaultAddress null.Int `json:"default_address"`
}

func (q *Queries) CreateUserAddress(ctx context.Context, arg CreateUserAddressParams) (UserAddress, error) {
	row := q.db.QueryRow(ctx, createUserAddress, arg.UserID, arg.AddressID, arg.DefaultAddress)
	var i UserAddress
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUserAddressWithAddress = `-- name: CreateUserAddressWithAddress :one
WITH t1 AS (
  INSERT INTO "address" as a (
    address_line,
    region,
    city
    ) VALUES (
    $1, $2, $3
    )
   RETURNING a.id, a.address_line, a.region, a.city
  ),

  t2 AS ( 
  INSERT INTO "user_address" (
    user_id,
    address_id,
    default_address
    ) VALUES 
    ( $4,
    (SELECT id FROM t1),
      $5
    ) 
  RETURNING user_id, address_id, default_address, created_at, updated_at
  )

SELECT 
user_id,
address_id,
default_address,
address_line,
region,
city 
FROM t1, t2
`

type CreateUserAddressWithAddressParams struct {
	AddressLine    string   `json:"address_line"`
	Region         string   `json:"region"`
	City           string   `json:"city"`
	UserID         int64    `json:"user_id"`
	DefaultAddress null.Int `json:"default_address"`
}

type CreateUserAddressWithAddressRow struct {
	UserID         int64    `json:"user_id"`
	AddressID      int64    `json:"address_id"`
	DefaultAddress null.Int `json:"default_address"`
	AddressLine    string   `json:"address_line"`
	Region         string   `json:"region"`
	City           string   `json:"city"`
}

func (q *Queries) CreateUserAddressWithAddress(ctx context.Context, arg CreateUserAddressWithAddressParams) (CreateUserAddressWithAddressRow, error) {
	row := q.db.QueryRow(ctx, createUserAddressWithAddress,
		arg.AddressLine,
		arg.Region,
		arg.City,
		arg.UserID,
		arg.DefaultAddress,
	)
	var i CreateUserAddressWithAddressRow
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.AddressLine,
		&i.Region,
		&i.City,
	)
	return i, err
}

const deleteUserAddress = `-- name: DeleteUserAddress :one
   
	

DELETE FROM "user_address"
WHERE user_id = $1
AND address_id = $2
RETURNING user_id, address_id, default_address, created_at, updated_at
`

type DeleteUserAddressParams struct {
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}

// -- name: UpdateUserAddressWithAddress :one
// WITH t1 AS (
//
//	    UPDATE "address" as a
//	    SET
//	    address_line = COALESCE(sqlc.narg(address_line),address_line),
//	    region = COALESCE(sqlc.narg(region),region),
//	    city= COALESCE(sqlc.narg(city),city)
//	    WHERE id = COALESCE(sqlc.arg(id),id)
//	    RETURNING a.id, a.address_line, a.region, a.city
//	   ),
//	    t2 AS (
//	    UPDATE "user_address"
//	    SET
//	    default_address = COALESCE(sqlc.narg(default_address),default_address)
//	    WHERE
//	    user_id = COALESCE(sqlc.arg(user_id),user_id)
//	    AND address_id = COALESCE(sqlc.arg(address_id),address_id)
//	    RETURNING user_id, address_id, default_address
//		)
//
// SELECT
// user_id,
// address_id,
// default_address,
// address_line,
// region,
// city From t1,t2;
func (q *Queries) DeleteUserAddress(ctx context.Context, arg DeleteUserAddressParams) (UserAddress, error) {
	row := q.db.QueryRow(ctx, deleteUserAddress, arg.UserID, arg.AddressID)
	var i UserAddress
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserAddress = `-- name: GetUserAddress :one
SELECT user_id, address_id, default_address, created_at, updated_at FROM "user_address"
WHERE user_id = $1
AND address_id = $2
LIMIT 1
`

type GetUserAddressParams struct {
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}

func (q *Queries) GetUserAddress(ctx context.Context, arg GetUserAddressParams) (UserAddress, error) {
	row := q.db.QueryRow(ctx, getUserAddress, arg.UserID, arg.AddressID)
	var i UserAddress
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserAddressWithAddress = `-- name: GetUserAddressWithAddress :one
SELECT ua.user_id, ua.address_id, ua.default_address, "address".address_line,  "address".region,  "address".city
FROM "user_address" AS ua 
JOIN "user" ON ua.user_id = "user".id 
JOIN "address" ON ua.address_id = "address".id
WHERE "user".id = $1
AND "address".id = $2
`

type GetUserAddressWithAddressParams struct {
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}

type GetUserAddressWithAddressRow struct {
	UserID         int64    `json:"user_id"`
	AddressID      int64    `json:"address_id"`
	DefaultAddress null.Int `json:"default_address"`
	AddressLine    string   `json:"address_line"`
	Region         string   `json:"region"`
	City           string   `json:"city"`
}

func (q *Queries) GetUserAddressWithAddress(ctx context.Context, arg GetUserAddressWithAddressParams) (GetUserAddressWithAddressRow, error) {
	row := q.db.QueryRow(ctx, getUserAddressWithAddress, arg.UserID, arg.AddressID)
	var i GetUserAddressWithAddressRow
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.AddressLine,
		&i.Region,
		&i.City,
	)
	return i, err
}

const listUserAddresses = `-- name: ListUserAddresses :many
SELECT user_id, address_id, default_address, created_at, updated_at FROM "user_address"
WHERE user_id = $1
ORDER BY address_id
LIMIT $2
OFFSET $3
`

type ListUserAddressesParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListUserAddresses(ctx context.Context, arg ListUserAddressesParams) ([]UserAddress, error) {
	rows, err := q.db.Query(ctx, listUserAddresses, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UserAddress{}
	for rows.Next() {
		var i UserAddress
		if err := rows.Scan(
			&i.UserID,
			&i.AddressID,
			&i.DefaultAddress,
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

const updateUserAddress = `-- name: UpdateUserAddress :one
UPDATE "user_address"
SET 
default_address = $1,
updated_at = now()
WHERE user_id = $2
AND address_id = $3
RETURNING user_id, address_id, default_address, created_at, updated_at
`

type UpdateUserAddressParams struct {
	DefaultAddress null.Int `json:"default_address"`
	UserID         int64    `json:"user_id"`
	AddressID      int64    `json:"address_id"`
}

func (q *Queries) UpdateUserAddress(ctx context.Context, arg UpdateUserAddressParams) (UserAddress, error) {
	row := q.db.QueryRow(ctx, updateUserAddress, arg.DefaultAddress, arg.UserID, arg.AddressID)
	var i UserAddress
	err := row.Scan(
		&i.UserID,
		&i.AddressID,
		&i.DefaultAddress,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
