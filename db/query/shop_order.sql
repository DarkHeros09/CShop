-- name: CreateShopOrder :one
INSERT INTO "shop_order" (
  order_number,
  user_id,
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

-- name: ListShopOrdersByUserID :many
SELECT os.status, 
(
  SELECT count(soi.id) FROM "shop_order_item" AS soi
  WHERE soi.order_id = so.id
) AS item_count,so.*
FROM "shop_order" AS so
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
WHERE so.user_id = $1
ORDER BY so.id DESC
LIMIT $2
OFFSET $3;

-- name: UpdateShopOrder :one
UPDATE "shop_order"
SET 
order_number = COALESCE(sqlc.narg(order_number),order_number),
user_id = COALESCE(sqlc.narg(user_id),user_id),
payment_method_id = COALESCE(sqlc.narg(payment_method_id),payment_method_id),
shipping_address_id = COALESCE(sqlc.narg(shipping_address_id),shipping_address_id),
order_total = COALESCE(sqlc.narg(order_total),order_total),
shipping_method_id = COALESCE(sqlc.narg(shipping_method_id),shipping_method_id),
order_status_id = COALESCE(sqlc.narg(order_status_id),order_status_id),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShopOrder :exec
DELETE FROM "shop_order"
WHERE id = $1;