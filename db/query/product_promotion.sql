-- name: CreateProductPromotion :one
INSERT INTO "product_promotion" (
  product_id,
  promotion_id,
  product_promotion_image,
  active
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: AdminCreateProductPromotion :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product_promotion" (
  product_id,
  promotion_id,
  product_promotion_image,
  active
)
SELECT sqlc.arg(product_id), sqlc.arg(promotion_id), sqlc.arg(product_promotion_image), sqlc.arg(active) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProductPromotion :one
SELECT * FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
LIMIT 1;

-- name: ListProductPromotionsWithImages :many
SELECT * FROM "product_promotion" AS pp
LEFT JOIN "product" AS p ON p.id = pp.product_id
JOIN "promotion" AS promo ON promo.id = pp.promotion_id AND promo.active = TRUE AND promo.start_date <= CURRENT_DATE AND promo.end_date >= CURRENT_DATE
JOIN "product_item" AS pi ON pi.product_id = p.id AND pi.active =TRUE
WHERE pp.product_promotion_image IS NOT NULL AND pp.active = TRUE
ORDER BY promo.start_date DESC;

-- name: ListProductPromotions :many
SELECT * FROM "product_promotion"
ORDER BY product_id
LIMIT $1
OFFSET $2;

-- name: AdminListProductPromotions :many
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT 
pp.product_id, p.name AS product_name, 
pp.promotion_id, promo.name AS promotion_name,
pp.product_promotion_image, pp.active FROM "product_promotion" AS pp
LEFT JOIN "product" AS p ON p.id = pp.product_id
LEFT JOIN "promotion" AS promo ON promo.id = pp.promotion_id
WHERE (SELECT is_admin FROM t1) = 1
ORDER BY product_id;

-- name: AdminUpdateProductPromotion :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "product_promotion"
SET
product_promotion_image = COALESCE(sqlc.narg(product_promotion_image),product_promotion_image),
active = COALESCE(sqlc.narg(active),active)
WHERE product_id = sqlc.arg(product_id)
AND promotion_id = sqlc.arg(promotion_id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: UpdateProductPromotion :one
UPDATE "product_promotion"
SET
product_promotion_image = COALESCE(sqlc.narg(product_promotion_image),product_promotion_image),
active = COALESCE(sqlc.narg(active),active)
WHERE product_id = sqlc.arg(product_id)
AND promotion_id = sqlc.arg(promotion_id)
RETURNING *;

-- name: DeleteProductPromotion :exec
DELETE FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
RETURNING *;