// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: user_session.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
)

const createUserSession = `-- name: CreateUserSession :one
INSERT INTO "user_session" (
  id,
  user_id,
  refresh_token,
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, user_id, refresh_token, user_agent, client_ip, is_blocked, created_at, updated_at, expires_at
`

type CreateUserSessionParams struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error) {
	row := q.db.QueryRow(ctx, createUserSession,
		arg.ID,
		arg.UserID,
		arg.RefreshToken,
		arg.UserAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiresAt,
	)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const getUserSession = `-- name: GetUserSession :one
SELECT id, user_id, refresh_token, user_agent, client_ip, is_blocked, created_at, updated_at, expires_at FROM "user_session"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserSession(ctx context.Context, id uuid.UUID) (UserSession, error) {
	row := q.db.QueryRow(ctx, getUserSession, id)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const updateUserSession = `-- name: UpdateUserSession :one
UPDATE "user_session"
SET 
is_blocked = COALESCE($1,is_blocked),
updated_at = now()
WHERE id = $2
AND user_id = $3
AND refresh_token = $4
RETURNING id, user_id, refresh_token, user_agent, client_ip, is_blocked, created_at, updated_at, expires_at
`

type UpdateUserSessionParams struct {
	IsBlocked    null.Bool `json:"is_blocked"`
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
}

func (q *Queries) UpdateUserSession(ctx context.Context, arg UpdateUserSessionParams) (UserSession, error) {
	row := q.db.QueryRow(ctx, updateUserSession,
		arg.IsBlocked,
		arg.ID,
		arg.UserID,
		arg.RefreshToken,
	)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}
