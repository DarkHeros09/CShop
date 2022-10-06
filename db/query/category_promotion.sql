-- name: CreateCategoryPromotion :one
INSERT INTO "category_promotion" (
  category_id,
  promotion_id,
  active
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetCategoryPromotion :one
SELECT * FROM "category_promotion"
WHERE category_id = $1 LIMIT 1;

-- name: ListCategoryPromotions :many
SELECT * FROM "category_promotion"
ORDER BY category_id
LIMIT $1
OFFSET $2;

-- name: UpdateCategoryPromotion :one
UPDATE "category_promotion"
SET
promotion_id = COALESCE(sqlc.narg(promotion_id),promotion_id),
active = COALESCE(sqlc.narg(active),active)
WHERE category_id = sqlc.arg(category_id)
RETURNING *;

-- name: DeleteCategoryPromotion :exec
DELETE FROM "category_promotion"
WHERE category_id = $1;