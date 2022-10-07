-- name: CreateShopOrder :one
INSERT INTO "shop_order" (
  user_id,
  order_date,
  payment_method_id,
  shipping_address_id,
  order_total,
  shipping_method_id,
  order_status_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetShopOrder :one
SELECT * FROM "shop_order"
WHERE id = $1 LIMIT 1;

-- name: ListShopOrders :many
SELECT * FROM "shop_order"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateShopOrder :one
UPDATE "shop_order"
SET 
user_id = COALESCE(sqlc.narg(user_id),user_id),
order_date = COALESCE(sqlc.narg(order_date),order_date),
payment_method_id = COALESCE(sqlc.narg(payment_method_id),payment_method_id),
shipping_address_id = COALESCE(sqlc.narg(shipping_address_id),shipping_address_id),
order_total = COALESCE(sqlc.narg(order_total),order_total),
shipping_method_id = COALESCE(sqlc.narg(shipping_method_id),shipping_method_id),
order_status_id = COALESCE(sqlc.narg(order_status_id),order_status_id)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShopOrder :exec
DELETE FROM "shop_order"
WHERE id = $1;