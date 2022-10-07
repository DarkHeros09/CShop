-- name: CreateShippingMethod :one
INSERT INTO "shipping_method" (
  name,
  price
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetShippingMethod :one
SELECT * FROM "shipping_method"
WHERE id = $1 LIMIT 1;

-- name: ListShippingMethods :many
SELECT * FROM "shipping_method"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateShippingMethod :one
UPDATE "shipping_method"
SET 
name = COALESCE(sqlc.narg(name),name),
price = COALESCE(sqlc.narg(price),price)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShippingMethod :exec
DELETE FROM "shipping_method"
WHERE id = $1;