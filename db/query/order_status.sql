-- name: CreateOrderStatus :one
INSERT INTO "order_status" (
  status
) VALUES (
  $1
)
ON CONFLICT(status) DO UPDATE SET status = $1
RETURNING *;

-- name: GetOrderStatus :one
SELECT * FROM "order_status"
WHERE id = $1 LIMIT 1;

-- name: GetOrderStatusByUserID :one
SELECT os.*, so.user_id
FROM "order_status" AS os
LEFT JOIN "shop_order" AS so ON so.order_status_id = os.id
WHERE so.user_id = $1
AND os.id = $2
LIMIT 1;

-- name: ListOrderStatuses :many
SELECT * FROM "order_status";
-- ORDER BY id
-- LIMIT $1
-- OFFSET $2;

-- name: ListOrderStatusesByUserID :many
SELECT os.*, so.user_id
FROM "order_status" AS os
LEFT JOIN "shop_order" AS so ON so.order_status_id = os.id
WHERE so.user_id = $3
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateOrderStatus :one
UPDATE "order_status"
SET 
status = COALESCE(sqlc.narg(status),status),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOrderStatus :exec
DELETE FROM "order_status"
WHERE id = $1;