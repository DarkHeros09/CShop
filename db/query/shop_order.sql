-- name: CreateShopOrder :one
INSERT INTO "shop_order" (
  track_number,
  order_number,
  user_id,
  -- payment_method_id,
  shipping_address_id,
  order_total,
  shipping_method_id,
  order_status_id
) VALUES (
  $1, 
  (
    SELECT COUNT(*) FROM "shop_order" so
    WHERE so.user_id = $2
     ) + 1, 
    $2, $3, $4, $5, $6
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
ROW_NUMBER() OVER(ORDER BY so.id) as order_number,
(
  SELECT COUNT(*) FROM "shop_order_item" AS soi
  WHERE soi.order_id = so.id
) AS item_count,so.*
FROM "shop_order" AS so
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
WHERE so.user_id = $1
ORDER BY so.id DESC
LIMIT $2
OFFSET $3;

-- name: ListShopOrdersByUserIDV2 :many
SELECT os.status,
-- ROW_NUMBER() OVER(ORDER BY so.id) AS order_number,
(
  SELECT COUNT(*) FROM "shop_order_item" AS soi
  WHERE soi.order_id = so.id
) AS item_count
, so.*, COUNT(*) OVER() AS total_count
FROM "shop_order" AS so
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
WHERE so.user_id = sqlc.arg(user_id)
AND CASE
WHEN COALESCE(sqlc.narg(order_status), '') != ''
THEN os.status = sqlc.narg(order_status)
    ELSE 1=1
END
ORDER BY so.id DESC
LIMIT $1;

-- name: ListShopOrdersByUserIDNextPage :many
SELECT os.status,
-- ROW_NUMBER() OVER(ORDER BY so.id) AS order_number,
(
  SELECT COUNT(*) FROM "shop_order_item" AS soi
  WHERE soi.order_id = so.id
) AS item_count
, so.*, COUNT(*) OVER() AS total_count
FROM "shop_order" AS so
LEFT JOIN "order_status" AS os ON os.id = so.order_status_id
WHERE so.user_id = sqlc.arg(user_id)
AND so.id < sqlc.arg(shop_order_id)
AND CASE
WHEN COALESCE(sqlc.narg(order_status), '') != ''
THEN os.status = sqlc.narg(order_status)
    ELSE 1=1
END
ORDER BY so.id DESC
LIMIT $1;

-- name: UpdateShopOrder :one
UPDATE "shop_order"
SET 
track_number = COALESCE(sqlc.narg(track_number),track_number),
user_id = COALESCE(sqlc.narg(user_id),user_id),
-- payment_method_id = COALESCE(sqlc.narg(payment_method_id),payment_method_id),
shipping_address_id = COALESCE(sqlc.narg(shipping_address_id),shipping_address_id),
order_total = COALESCE(sqlc.narg(order_total),order_total),
shipping_method_id = COALESCE(sqlc.narg(shipping_method_id),shipping_method_id),
order_status_id = COALESCE(sqlc.narg(order_status_id),order_status_id),
updated_at = NOW(),
completed_at = CASE
    WHEN order_status_id != 2 AND sqlc.narg(order_status_id) =2
    THEN NOW()
    ELSE completed_at
END
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteShopOrder :exec
DELETE FROM "shop_order"
WHERE id = $1;

-- name: GetShopOrdersCountByStatusId :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(*) AS orders_count FROM shop_order
WHERE order_status_id = sqlc.arg(order_status_id)
AND EXISTS(SELECT is_admin FROM t1);

-- name: GetTotalShopOrder :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(*) AS orders_count FROM shop_order
WHERE EXISTS(SELECT is_admin FROM t1);

-- name: GetCompletedDailyOrderTotal :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT
    CAST(SUM(CAST(order_total AS NUMERIC))AS VARCHAR) AS daily_revenue
FROM
    shop_order
WHERE order_status_id = 2
AND updated_at >= CURRENT_DATE
AND updated_at < CURRENT_DATE + INTERVAL '1 day'
AND EXISTS(SELECT is_admin FROM t1);