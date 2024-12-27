// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: notification.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const createNotification = `-- name: CreateNotification :one
INSERT INTO "notification" (
  user_id,
  device_id,
fcm_token
) VALUES (
  $1, $2, $3
) 
RETURNING user_id, device_id, fcm_token, created_at, updated_at
`

type CreateNotificationParams struct {
	UserID   int64       `json:"user_id"`
	DeviceID null.String `json:"device_id"`
	FcmToken null.String `json:"fcm_token"`
}

// ON CONFLICT(user_id) DO UPDATE SET
// device_id = EXCLUDED.device_id,
// fcm_token = EXCLUDED.fcm_token
func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, createNotification, arg.UserID, arg.DeviceID, arg.FcmToken)
	var i Notification
	err := row.Scan(
		&i.UserID,
		&i.DeviceID,
		&i.FcmToken,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteNotification = `-- name: DeleteNotification :one
DELETE FROM "notification"
WHERE user_id = $1
AND device_id = $2
RETURNING user_id, device_id, fcm_token, created_at, updated_at
`

type DeleteNotificationParams struct {
	UserID   int64       `json:"user_id"`
	DeviceID null.String `json:"device_id"`
}

func (q *Queries) DeleteNotification(ctx context.Context, arg DeleteNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, deleteNotification, arg.UserID, arg.DeviceID)
	var i Notification
	err := row.Scan(
		&i.UserID,
		&i.DeviceID,
		&i.FcmToken,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteNotificationAllByUser = `-- name: DeleteNotificationAllByUser :exec
DELETE FROM "notification"
WHERE user_id = $1
`

func (q *Queries) DeleteNotificationAllByUser(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, deleteNotificationAllByUser, userID)
	return err
}

const getNotification = `-- name: GetNotification :one
SELECT user_id, device_id, fcm_token, created_at, updated_at FROM "notification"
WHERE user_id = $1
AND device_id = $2
`

type GetNotificationParams struct {
	UserID   int64       `json:"user_id"`
	DeviceID null.String `json:"device_id"`
}

func (q *Queries) GetNotification(ctx context.Context, arg GetNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, getNotification, arg.UserID, arg.DeviceID)
	var i Notification
	err := row.Scan(
		&i.UserID,
		&i.DeviceID,
		&i.FcmToken,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNotificationV2 = `-- name: GetNotificationV2 :one
SELECT user_id, device_id, fcm_token, created_at, updated_at FROM "notification"
WHERE user_id = $1
ORDER BY updated_at DESC, created_at DESC
LIMIT 1
`

func (q *Queries) GetNotificationV2(ctx context.Context, userID int64) (Notification, error) {
	row := q.db.QueryRow(ctx, getNotificationV2, userID)
	var i Notification
	err := row.Scan(
		&i.UserID,
		&i.DeviceID,
		&i.FcmToken,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateNotification = `-- name: UpdateNotification :one
UPDATE "notification"
SET 
fcm_token = COALESCE($1,fcm_token),
updated_at = now()
WHERE user_id = $2
AND device_id = $3
RETURNING user_id, device_id, fcm_token, created_at, updated_at
`

type UpdateNotificationParams struct {
	FcmToken null.String `json:"fcm_token"`
	UserID   int64       `json:"user_id"`
	DeviceID null.String `json:"device_id"`
}

func (q *Queries) UpdateNotification(ctx context.Context, arg UpdateNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, updateNotification, arg.FcmToken, arg.UserID, arg.DeviceID)
	var i Notification
	err := row.Scan(
		&i.UserID,
		&i.DeviceID,
		&i.FcmToken,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
