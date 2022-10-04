-- name: CreateProduct :one
INSERT INTO "product" (
  category_id,
  name,
  description,
  product_image,
  active
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetProduct :one
SELECT * FROM "product"
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM "product"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateProduct :one
UPDATE "product"
SET
category_id = COALESCE(sqlc.narg(category_id),category_id),
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description),
product_image = COALESCE(sqlc.narg(product_image),product_image),
active = COALESCE(sqlc.narg(active),active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM "product"
WHERE id = $1;