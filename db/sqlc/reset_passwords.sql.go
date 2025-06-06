// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: reset_passwords.sql

package db

import (
	"context"
	"time"
)

const createResetPassword = `-- name: CreateResetPassword :one
INSERT INTO "reset_passwords" (
    user_id,
    -- email,
    secret_code
) VALUES (
    $1, $2
) RETURNING id, user_id, secret_code, is_used, created_at, updated_at, expired_at
`

type CreateResetPasswordParams struct {
	UserID     int64  `json:"user_id"`
	SecretCode string `json:"secret_code"`
}

func (q *Queries) CreateResetPassword(ctx context.Context, arg CreateResetPasswordParams) (ResetPassword, error) {
	row := q.db.QueryRow(ctx, createResetPassword, arg.UserID, arg.SecretCode)
	var i ResetPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SecretCode,
		&i.IsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiredAt,
	)
	return i, err
}

const getLastUsedResetPassword = `-- name: GetLastUsedResetPassword :one
SELECT rp.id, rp.user_id, rp.secret_code, rp.is_used, rp.created_at, rp.updated_at, rp.expired_at FROM "reset_passwords" AS rp
JOIN "user" AS u ON u.id = rp.user_id
WHERE u.email = $1
AND is_used = TRUE
AND expired_at > now()
ORDER BY rp.updated_at DESC, rp.created_at DESC
LIMIT 1
`

// AND secret_code = $2
func (q *Queries) GetLastUsedResetPassword(ctx context.Context, email string) (ResetPassword, error) {
	row := q.db.QueryRow(ctx, getLastUsedResetPassword, email)
	var i ResetPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SecretCode,
		&i.IsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiredAt,
	)
	return i, err
}

const getResetPassword = `-- name: GetResetPassword :one
SELECT id, user_id, secret_code, is_used, created_at, updated_at, expired_at FROM "reset_passwords"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetResetPassword(ctx context.Context, id int64) (ResetPassword, error) {
	row := q.db.QueryRow(ctx, getResetPassword, id)
	var i ResetPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SecretCode,
		&i.IsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiredAt,
	)
	return i, err
}

const getResetPasswordUserIDByID = `-- name: GetResetPasswordUserIDByID :one
SELECT user_id FROM "reset_passwords"
WHERE
    id = $1
    AND secret_code = $2
    AND is_used = FALSE
    AND expired_at > now()
LIMIT 1
`

type GetResetPasswordUserIDByIDParams struct {
	ID         int64  `json:"id"`
	SecretCode string `json:"secret_code"`
}

func (q *Queries) GetResetPasswordUserIDByID(ctx context.Context, arg GetResetPasswordUserIDByIDParams) (int64, error) {
	row := q.db.QueryRow(ctx, getResetPasswordUserIDByID, arg.ID, arg.SecretCode)
	var user_id int64
	err := row.Scan(&user_id)
	return user_id, err
}

const getResetPasswordsByEmail = `-- name: GetResetPasswordsByEmail :one
SELECT u.email, u.username, u.is_blocked AS is_blocked_user, u.is_email_verified, 
rp.id, rp.user_id, rp.secret_code, rp.is_used, rp.created_at, rp.updated_at, rp.expired_at FROM "reset_passwords" AS rp
JOIN "user" AS u ON u.id = rp.user_id
WHERE u.email = $1
ORDER BY rp.created_at DESC
LIMIT 1
`

type GetResetPasswordsByEmailRow struct {
	Email           string    `json:"email"`
	Username        string    `json:"username"`
	IsBlockedUser   bool      `json:"is_blocked_user"`
	IsEmailVerified bool      `json:"is_email_verified"`
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	SecretCode      string    `json:"secret_code"`
	IsUsed          bool      `json:"is_used"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiredAt       time.Time `json:"expired_at"`
}

func (q *Queries) GetResetPasswordsByEmail(ctx context.Context, email string) (GetResetPasswordsByEmailRow, error) {
	row := q.db.QueryRow(ctx, getResetPasswordsByEmail, email)
	var i GetResetPasswordsByEmailRow
	err := row.Scan(
		&i.Email,
		&i.Username,
		&i.IsBlockedUser,
		&i.IsEmailVerified,
		&i.ID,
		&i.UserID,
		&i.SecretCode,
		&i.IsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiredAt,
	)
	return i, err
}

const updateResetPassword = `-- name: UpdateResetPassword :one
UPDATE "reset_passwords"
SET
    is_used = TRUE
WHERE
    id = $1
    AND secret_code = $2
    AND is_used = FALSE
    AND expired_at > now()
RETURNING id, user_id, secret_code, is_used, created_at, updated_at, expired_at
`

type UpdateResetPasswordParams struct {
	ID         int64  `json:"id"`
	SecretCode string `json:"secret_code"`
}

func (q *Queries) UpdateResetPassword(ctx context.Context, arg UpdateResetPasswordParams) (ResetPassword, error) {
	row := q.db.QueryRow(ctx, updateResetPassword, arg.ID, arg.SecretCode)
	var i ResetPassword
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SecretCode,
		&i.IsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiredAt,
	)
	return i, err
}
