-- name: CreatePromotion :one
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
RETURNING *;

-- name: GetPromotion :one
SELECT * FROM "promotion"
WHERE id = $1 LIMIT 1;

-- name: ListPromotions :many
SELECT * FROM "promotion"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdatePromotion :one
UPDATE "promotion"
SET
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description),
discount_rate = COALESCE(sqlc.narg(discount_rate),discount_rate),
active = COALESCE(sqlc.narg(active),active),
start_date = COALESCE(sqlc.narg(start_date),start_date),
end_date = COALESCE(sqlc.narg(end_date),end_date)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePromotion :exec
DELETE FROM "promotion"
WHERE id = $1;