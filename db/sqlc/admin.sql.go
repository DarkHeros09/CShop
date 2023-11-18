// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: admin.sql

package db

import (
	"context"
)

const createAdmin = `-- name: CreateAdmin :one
INSERT INTO "admin" (
  username,
  email,
  password,
  type_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, username, email, password, active, type_id, created_at, updated_at, last_login
`

type CreateAdminParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	TypeID   int64  `json:"type_id"`
}

func (q *Queries) CreateAdmin(ctx context.Context, arg CreateAdminParams) (Admin, error) {
	row := q.db.QueryRow(ctx, createAdmin,
		arg.Username,
		arg.Email,
		arg.Password,
		arg.TypeID,
	)
	var i Admin
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Active,
		&i.TypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLogin,
	)
	return i, err
}

const deleteAdmin = `-- name: DeleteAdmin :exec
DELETE FROM "admin"
WHERE id = $1
`

func (q *Queries) DeleteAdmin(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteAdmin, id)
	return err
}

const getAdmin = `-- name: GetAdmin :one
SELECT id, username, email, password, active, type_id, created_at, updated_at, last_login FROM "admin"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetAdmin(ctx context.Context, id int64) (Admin, error) {
	row := q.db.QueryRow(ctx, getAdmin, id)
	var i Admin
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Active,
		&i.TypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLogin,
	)
	return i, err
}

const getAdminByEmail = `-- name: GetAdminByEmail :one
SELECT id, username, email, password, active, type_id, created_at, updated_at, last_login FROM "admin"
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetAdminByEmail(ctx context.Context, email string) (Admin, error) {
	row := q.db.QueryRow(ctx, getAdminByEmail, email)
	var i Admin
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Active,
		&i.TypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLogin,
	)
	return i, err
}

const listAdmins = `-- name: ListAdmins :many
SELECT id, username, email, password, active, type_id, created_at, updated_at, last_login FROM "admin"
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListAdminsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListAdmins(ctx context.Context, arg ListAdminsParams) ([]Admin, error) {
	rows, err := q.db.Query(ctx, listAdmins, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Admin{}
	for rows.Next() {
		var i Admin
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.Password,
			&i.Active,
			&i.TypeID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LastLogin,
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

const updateAdmin = `-- name: UpdateAdmin :one
UPDATE "admin"
SET active = $2,
updated_at = now()
WHERE id = $1
RETURNING id, username, email, password, active, type_id, created_at, updated_at, last_login
`

type UpdateAdminParams struct {
	ID     int64 `json:"id"`
	Active bool  `json:"active"`
}

func (q *Queries) UpdateAdmin(ctx context.Context, arg UpdateAdminParams) (Admin, error) {
	row := q.db.QueryRow(ctx, updateAdmin, arg.ID, arg.Active)
	var i Admin
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Active,
		&i.TypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLogin,
	)
	return i, err
}
