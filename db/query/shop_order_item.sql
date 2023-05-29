-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  quantity,
  price
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetShopOrderItem :one
SELECT * FROM "shop_order_item"
WHERE id = $1 LIMIT 1;

-- name: GetShopOrderItemByUserIDOrderID :one
SELECT soi.*, so.user_id
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
WHERE so.user_id = $1
AND soi.order_id = $2 
LIMIT 1;

-- name: ListShopOrderItems :many
SELECT * FROM "shop_order_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListShopOrderItemsByUserIDOrderID :many
-- SELECT * FROM "shop_order_item"
-- WHERE order_id = $1
-- ORDER BY id;
SELECT soi.*
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
WHERE so.user_id = $1
AND soi.order_id = $2
ORDER BY soi.id;

-- name: ListShopOrderItemsByUserID :many
SELECT so.*, soi.* 
FROM "shop_order" AS so
LEFT JOIN "shop_order_item" AS soi ON soi.order_id = so.id
WHERE so.user_id = $3
ORDER BY so.id
LIMIT $1
OFFSET $2;

-- -- name: ListShopOrderItemsByOrderID :many
-- SELECT * FROM "shop_order_item"
-- WHERE order_id = $1
-- ORDER BY id;

-- name: UpdateShopOrderItem :one
UPDATE "shop_order_item"
SET 
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id),
order_id = COALESCE(sqlc.narg(order_id),order_id),
quantity = COALESCE(sqlc.narg(quantity),quantity),
price = COALESCE(sqlc.narg(price),price),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShopOrderItem :exec
DELETE FROM "shop_order_item"
WHERE id = $1;