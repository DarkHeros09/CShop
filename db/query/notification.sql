-- name: CreateNotification :one
INSERT INTO "notification" (
  user_id,
  device_id,
fcm_token
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetNotification :one
SELECT * FROM "notification"
WHERE user_id = $1
AND device_id = $2;

-- name: UpdateNotification :one
UPDATE "notification"
SET 
fcm_token = COALESCE(sqlc.narg(fcm_token),fcm_token),
updated_at = now()
WHERE user_id = sqlc.arg(user_id)
AND device_id = sqlc.arg(device_id)
RETURNING *;

-- name: DeleteNotification :one
DELETE FROM "notification"
WHERE user_id = sqlc.arg(user_id)
AND device_id = sqlc.arg(device_id)
RETURNING *;

-- name: DeleteNotificationAllByUser :exec
DELETE FROM "notification"
WHERE user_id = sqlc.arg(user_id);