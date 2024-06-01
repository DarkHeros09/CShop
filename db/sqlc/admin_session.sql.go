// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: admin_session.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	null "github.com/guregu/null/v5"
)

const createAdminSession = `-- name: CreateAdminSession :one
INSERT INTO "admin_session" (
  id,
  admin_id,
  refresh_token,
  admin_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, admin_id, refresh_token, admin_agent, client_ip, is_blocked, created_at, updated_at, expires_at
`

type CreateAdminSessionParams struct {
	ID           uuid.UUID `json:"id"`
	AdminID      int64     `json:"admin_id"`
	RefreshToken string    `json:"refresh_token"`
	AdminAgent   string    `json:"admin_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (q *Queries) CreateAdminSession(ctx context.Context, arg CreateAdminSessionParams) (AdminSession, error) {
	row := q.db.QueryRow(ctx, createAdminSession,
		arg.ID,
		arg.AdminID,
		arg.RefreshToken,
		arg.AdminAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiresAt,
	)
	var i AdminSession
	err := row.Scan(
		&i.ID,
		&i.AdminID,
		&i.RefreshToken,
		&i.AdminAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const getAdminSession = `-- name: GetAdminSession :one
SELECT id, admin_id, refresh_token, admin_agent, client_ip, is_blocked, created_at, updated_at, expires_at FROM "admin_session"
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetAdminSession(ctx context.Context, id uuid.UUID) (AdminSession, error) {
	row := q.db.QueryRow(ctx, getAdminSession, id)
	var i AdminSession
	err := row.Scan(
		&i.ID,
		&i.AdminID,
		&i.RefreshToken,
		&i.AdminAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const updateAdminSession = `-- name: UpdateAdminSession :one
UPDATE "admin_session"
SET 
is_blocked = COALESCE($1,is_blocked),
updated_at = now()
WHERE id = $2
AND admin_id = $3
AND refresh_token = $4
RETURNING id, admin_id, refresh_token, admin_agent, client_ip, is_blocked, created_at, updated_at, expires_at
`

type UpdateAdminSessionParams struct {
	IsBlocked    null.Bool `json:"is_blocked"`
	ID           uuid.UUID `json:"id"`
	AdminID      int64     `json:"admin_id"`
	RefreshToken string    `json:"refresh_token"`
}

func (q *Queries) UpdateAdminSession(ctx context.Context, arg UpdateAdminSessionParams) (AdminSession, error) {
	row := q.db.QueryRow(ctx, updateAdminSession,
		arg.IsBlocked,
		arg.ID,
		arg.AdminID,
		arg.RefreshToken,
	)
	var i AdminSession
	err := row.Scan(
		&i.ID,
		&i.AdminID,
		&i.RefreshToken,
		&i.AdminAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ExpiresAt,
	)
	return i, err
}
