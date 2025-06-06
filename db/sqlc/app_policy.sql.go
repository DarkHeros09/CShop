// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: app_policy.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const createAppPolicy = `-- name: CreateAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $2
    AND active = TRUE
    )
INSERT INTO "app_policy" 
( "policy" )
SELECT $1 FROM t1
WHERE EXISTS (SELECT 1 FROM t1)
RETURNING id, policy, created_at, updated_at
`

type CreateAppPolicyParams struct {
	Policy  null.String `json:"policy"`
	AdminID int64       `json:"admin_id"`
}

func (q *Queries) CreateAppPolicy(ctx context.Context, arg CreateAppPolicyParams) (AppPolicy, error) {
	row := q.db.QueryRow(ctx, createAppPolicy, arg.Policy, arg.AdminID)
	var i AppPolicy
	err := row.Scan(
		&i.ID,
		&i.Policy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAppPolicy = `-- name: DeleteAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $2
    AND active = TRUE
    )
DELETE FROM "app_policy"
WHERE "app_policy".id = $1
AND EXISTS (SELECT 1 FROM t1)
RETURNING id, policy, created_at, updated_at
`

type DeleteAppPolicyParams struct {
	ID      int64 `json:"id"`
	AdminID int64 `json:"admin_id"`
}

func (q *Queries) DeleteAppPolicy(ctx context.Context, arg DeleteAppPolicyParams) (AppPolicy, error) {
	row := q.db.QueryRow(ctx, deleteAppPolicy, arg.ID, arg.AdminID)
	var i AppPolicy
	err := row.Scan(
		&i.ID,
		&i.Policy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAppPolicy = `-- name: GetAppPolicy :one
SELECT id, policy, created_at, updated_at FROM "app_policy"
LIMIT 1
`

func (q *Queries) GetAppPolicy(ctx context.Context) (AppPolicy, error) {
	row := q.db.QueryRow(ctx, getAppPolicy)
	var i AppPolicy
	err := row.Scan(
		&i.ID,
		&i.Policy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateAppPolicy = `-- name: UpdateAppPolicy :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = $3
    AND active = TRUE
    )
UPDATE "app_policy"
SET 
policy = COALESCE($1,policy),
updated_at = NOW()
WHERE "app_policy".id = $2
AND EXISTS (SELECT 1 FROM t1)
RETURNING id, policy, created_at, updated_at
`

type UpdateAppPolicyParams struct {
	Policy  null.String `json:"policy"`
	ID      int64       `json:"id"`
	AdminID int64       `json:"admin_id"`
}

func (q *Queries) UpdateAppPolicy(ctx context.Context, arg UpdateAppPolicyParams) (AppPolicy, error) {
	row := q.db.QueryRow(ctx, updateAppPolicy, arg.Policy, arg.ID, arg.AdminID)
	var i AppPolicy
	err := row.Scan(
		&i.ID,
		&i.Policy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
