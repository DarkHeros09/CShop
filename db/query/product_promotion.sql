-- name: CreateProductPromotion :one
INSERT INTO "product_promotion" (
  product_id,
  promotion_id,
  active
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetProductPromotion :one
SELECT * FROM "product_promotion"
WHERE product_id = sqlc.arg(product_id)
AND promotion_id = sqlc.arg(promotion_id)
LIMIT 1;

-- name: ListProductPromotions :many
SELECT * FROM "product_promotion"
ORDER BY product_id
LIMIT $1
OFFSET $2;

-- name: UpdateProductPromotion :one
UPDATE "product_promotion"
SET
active = COALESCE(sqlc.narg(active),active)
WHERE product_id = sqlc.arg(product_id)
AND promotion_id = sqlc.arg(promotion_id)
RETURNING *;

-- name: DeleteProductPromotion :exec
DELETE FROM "product_promotion"
WHERE product_id = $1
AND promotion_id = $2
RETURNING *;