-- name: CreateBrandPromotion :one
INSERT INTO "brand_promotion" (
  brand_id,
  promotion_id,
  brand_promotion_image,
  active
) VALUES (
  $1, $2, $3, $4
) ON CONFLICT(brand_id) DO UPDATE SET 
promotion_id = EXCLUDED.promotion_id,
brand_promotion_image = EXCLUDED.brand_promotion_image,
active = EXCLUDED.active
RETURNING *;

-- name: GetBrandPromotion :one
SELECT * FROM "brand_promotion"
WHERE brand_id = $1
AND promotion_id = $2 
LIMIT 1;

-- name: ListBrandPromotionsWithImages :many
SELECT * FROM "brand_promotion" AS bp
LEFT JOIN "product_brand" AS pb ON pb.id = bp.brand_id
JOIN "promotion" AS promo ON promo.id = bp.promotion_id AND promo.active = true AND promo.start_date <= CURRENT_DATE AND promo.end_date >= CURRENT_DATE
WHERE bp.brand_promotion_image IS NOT NULL AND bp.active = true;

-- name: ListBrandPromotions :many
SELECT * FROM "brand_promotion"
ORDER BY brand_id
LIMIT $1
OFFSET $2;

-- name: UpdateBrandPromotion :one
UPDATE "brand_promotion"
SET
brand_promotion_image = COALESCE(sqlc.narg(brand_promotion_image),brand_promotion_image),
active = COALESCE(sqlc.narg(active),active)
WHERE brand_id = sqlc.arg(brand_id)
AND promotion_id = sqlc.arg(promotion_id)
RETURNING *;

-- name: DeleteBrandPromotion :exec
DELETE FROM "brand_promotion"
WHERE brand_id = $1
AND promotion_id = $2
RETURNING *;