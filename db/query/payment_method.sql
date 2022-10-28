-- name: CreatePaymentMethod :one
INSERT INTO "payment_method" (
  user_id,
  payment_type_id,
  provider
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetPaymentMethod :one
SELECT * FROM "payment_method"
WHERE id = $1 LIMIT 1;

-- name: ListPaymentMethods :many
SELECT * FROM "payment_method"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdatePaymentMethod :one
UPDATE "payment_method"
SET 
user_id = COALESCE(sqlc.narg(user_id),user_id),
payment_type_id = COALESCE(sqlc.narg(payment_type_id),payment_type_id),
provider = COALESCE(sqlc.narg(provider),provider)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePaymentMethod :exec
DELETE FROM "payment_method"
WHERE id = $1;