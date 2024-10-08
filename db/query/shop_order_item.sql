-- name: CreateShopOrderItem :one
INSERT INTO "shop_order_item" (
  product_item_id,
  order_id,
  quantity,
  price,
  discount,
  shipping_method_price
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
SELECT os.status, so.track_number, soi.shipping_method_price AS delivery_price, so.order_total, soi.*, p.name AS product_name, pt.value as payment_type,
-- pi.product_image, 
pimg.product_image_1 AS product_image,
pcolor.color_value AS product_color, psize.size_value AS product_size,
pi.active AS product_active, a.address_line, a.region, a.city,
DENSE_RANK() OVER(ORDER BY so.id) as order_number
-- , pt.value AS payment_type 
FROM "shop_order_item" AS soi
LEFT JOIN "shop_order" AS so ON so.id = soi.order_id
LEFT JOIN "product_item" AS pi ON pi.id = soi.product_item_id
LEFT JOIN "product_image" AS pimg ON pimg.id = pi.image_id
LEFT JOIN "product_size" AS psize ON psize.id = pi.size_id
LEFT JOIN "product_color" AS pcolor ON pcolor.id = pi.color_id
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
LEFT JOIN "product" AS p ON p.id = pi.product_id
LEFT JOIN "address" AS a ON a.id = so.shipping_address_id
-- LEFT JOIN "payment_method" AS pm ON pm.id = so.payment_method_id
LEFT JOIN "payment_type" AS pt ON pt.id = so.payment_type_id
-- LEFT JOIN "shipping_method" AS sm ON sm.id = so.shipping_method_id
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
quantity = COALESCE(sqlc.narg(quantity),quantity),
price = COALESCE(sqlc.narg(price),price),
discount = COALESCE(sqlc.narg(discount),discount),
shipping_method_price = COALESCE(sqlc.narg(shipping_method_price),shipping_method_price),
updated_at = now()
WHERE id = sqlc.arg(id)
AND order_id = sqlc.arg(order_id)
AND product_item_id = sqlc.arg(product_item_id)
RETURNING *;

-- name: DeleteShopOrderItem :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "shop_order_item"
WHERE "shop_order_item".id = $1
AND (SELECT is_admin FROM t1) = 1
RETURNING *;