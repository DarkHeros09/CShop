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

-- name: ListShopOrderItems :many
SELECT * FROM "shop_order_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateShopOrderItem :one
UPDATE "shop_order_item"
SET 
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id),
order_id = COALESCE(sqlc.narg(order_id),order_id),
quantity = COALESCE(sqlc.narg(quantity),quantity),
price = COALESCE(sqlc.narg(price),price)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShopOrderItem :exec
DELETE FROM "shop_order_item"
WHERE id = $1;