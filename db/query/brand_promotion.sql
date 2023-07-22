-- name: CreateBrandPromotion :one
INSERT INTO "brand_promotion" (
  brand_id,
  promotion_id,
  active
) VALUES (
  $1, $2, $3
) ON CONFLICT(brand_id) DO UPDATE SET 
promotion_id = EXCLUDED.promotion_id,
active = EXCLUDED.active
RETURNING *;

-- name: GetBrandPromotion :one
SELECT * FROM "brand_promotion"
WHERE brand_id = $1
AND promotion_id = $2 
LIMIT 1;

-- name: ListBrandPromotions :many
SELECT * FROM "brand_promotion"
ORDER BY brand_id
LIMIT $1
OFFSET $2;

-- name: UpdateBrandPromotion :one
UPDATE "brand_promotion"
SET
active = COALESCE(sqlc.narg(active),active)
WHERE brand_id = sqlc.arg(brand_id)
AND promotion_id = sqlc.arg(promotion_id)
RETURNING *;

-- name: DeleteBrandPromotion :exec
DELETE FROM "brand_promotion"
WHERE brand_id = $1
AND promotion_id = $2
RETURNING *;