-- name: CreatePaymentType :one
INSERT INTO "payment_type" (
  value
) VALUES (
  $1
) 
ON CONFLICT(value) DO UPDATE SET value = $1
RETURNING *;

-- name: GetPaymentType :one
SELECT * FROM "payment_type"
WHERE id = $1 LIMIT 1;

-- name: ListPaymentTypes :many
SELECT * FROM "payment_type";
-- ORDER BY id
-- LIMIT $1
-- OFFSET $2;

-- name: UpdatePaymentType :one
UPDATE "payment_type"
SET 
value = COALESCE(sqlc.narg(value),value),
is_active = COALESCE(sqlc.narg(is_active),is_active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePaymentType :exec
DELETE FROM "payment_type"
WHERE id = $1;