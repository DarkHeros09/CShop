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

-- name: GetCategoryPromotion :one
SELECT * FROM "category_promotion"
WHERE category_id = $1
AND promotion_id = $2 
LIMIT 1;

-- name: ListCategoryPromotionsWithImages :many
SELECT * FROM "category_promotion" AS cp
LEFT JOIN "product_category" AS pc ON pc.id = cp.category_id
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