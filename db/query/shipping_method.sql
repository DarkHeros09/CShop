-- name: CreateShippingMethod :one
INSERT INTO "shipping_method" (
  name,
  price
) VALUES (
  $1, $2
)
ON CONFLICT (name) DO UPDATE SET 
name = EXCLUDED.name,
price = EXCLUDED.price
RETURNING *;

-- name: GetShippingMethod :one
SELECT * FROM "shipping_method"
WHERE id = $1 LIMIT 1;

-- name: GetShippingMethodByUserID :one
SELECT sm.*, so.user_id
FROM "shipping_method" AS sm
LEFT JOIN "shop_order" AS so ON so.shipping_method_id = sm.id
WHERE so.user_id = $1
AND sm.id = $2
LIMIT 1;

-- name: ListShippingMethods :many
SELECT * FROM "shipping_method";
-- ORDER BY id
-- LIMIT $1
-- OFFSET $2;

-- name: ListShippingMethodsByUserID :many
SELECT sm.*, so.user_id
FROM "shipping_method" AS sm
LEFT JOIN "shop_order" AS so ON so.shipping_method_id = sm.id
WHERE so.user_id = $3
ORDER BY sm.id
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