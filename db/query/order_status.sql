-- name: CreateOrderStatus :one
INSERT INTO "order_status" (
  status
) VALUES (
  $1
)
RETURNING *;

-- name: GetOrderStatus :one
SELECT * FROM "order_status"
WHERE id = $1 LIMIT 1;

-- name: ListOrderStatuses :many
SELECT * FROM "order_status"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateOrderStatus :one
UPDATE "order_status"
SET 
status = COALESCE(sqlc.narg(status),status)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOrderStatus :exec
DELETE FROM "order_status"
WHERE id = $1;