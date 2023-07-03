-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  size,
  color,
  quantity,
  price
) VALUES (
  $1, $2, $3, $4, $5, $6
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
SELECT os.status, so.track_number, sm.price AS delivery_price, so.order_total, soi.*, p.name AS product_name,
-- pi.product_image, 
pimg.product_image_1 AS product_image,
pi.active AS product_active, a.address_line, a.region, a.city,
DENSE_RANK() OVER(ORDER BY so.id) as order_number, pt.value AS payment_type 
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
LEFT JOIN "product_item" AS pi ON pi.id = soi.product_item_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "address" AS a ON a.id = so.shipping_address_id
LEFT JOIN "payment_method" AS pm ON pm.id = so.payment_method_id
LEFT JOIN "payment_type" AS pt ON pt.id = pm.payment_type_id
LEFT JOIN "shipping_method" AS sm ON sm.id = so.shipping_method_id
WHERE so.user_id = $1
AND soi.order_id = $2;
-- ORDER BY soi.id;

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