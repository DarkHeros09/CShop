-- name: CreateCategoryPromotion :one
INSERT INTO "category_promotion" (
  category_id,
  promotion_id,
  category_promotion_image,
  active
) VALUES (
  $1, $2, $3, $4
) ON CONFLICT(category_id) DO UPDATE SET 
promotion_id = EXCLUDED.promotion_id,
category_promotion_image = EXCLUDED.category_promotion_image,
active = EXCLUDED.active
RETURNING *;

-- name: AdminCreateCategoryPromotion :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "category_promotion" (
  category_id,
  promotion_id,
  category_promotion_image,
  active
)
SELECT sqlc.arg(category_id), sqlc.arg(promotion_id), sqlc.arg(category_promotion_image), sqlc.arg(active) FROM t1
WHERE is_admin=1
ON CONFLICT(category_id) DO UPDATE SET 
promotion_id = EXCLUDED.promotion_id,
category_promotion_image = EXCLUDED.category_promotion_image,
active = EXCLUDED.active
RETURNING *;

-- name: GetCategoryPromotion :one
SELECT * FROM "category_promotion"
WHERE category_id = $1
AND promotion_id = $2 
LIMIT 1;

-- name: ListCategoryPromotionsWithImages :many
SELECT * FROM "category_promotion" AS cp
LEFT JOIN "product_category" AS pc ON pc.id = cp.category_id
JOIN "promotion" AS promo ON promo.id = cp.promotion_id AND promo.active = true AND promo.start_date <= CURRENT_DATE AND promo.end_date >= CURRENT_DATE
WHERE cp.category_promotion_image IS NOT NULL AND cp.active = true;

-- name: ListCategoryPromotions :many
SELECT * FROM "category_promotion"
ORDER BY category_id
LIMIT $1
OFFSET $2;

-- name: UpdateCategoryPromotion :one
UPDATE "category_promotion"
SET
category_promotion_image = COALESCE(sqlc.narg(category_promotion_image),category_promotion_image),
active = COALESCE(sqlc.narg(active),active)
WHERE category_id = sqlc.arg(category_id)
AND promotion_id = sqlc.arg(promotion_id)
RETURNING *;

-- name: DeleteCategoryPromotion :exec
DELETE FROM "category_promotion"
WHERE category_id = $1
AND promotion_id = $2
RETURNING *;