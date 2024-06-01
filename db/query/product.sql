-- name: CreateProduct :one
INSERT INTO "product" (
  category_id,
  brand_id,
  name,
  description,
  -- product_image,
  active
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: AdminCreateProduct :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
INSERT INTO "product" (
  category_id,
  brand_id,
  name,
  description,
  active
)
SELECT sqlc.arg(category_id), sqlc.arg(brand_id), sqlc.arg(name), sqlc.arg(description), sqlc.arg(active) FROM t1
WHERE is_admin=1
RETURNING *;

-- name: GetProduct :one
SELECT * FROM "product"
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
-- WITH total_records AS (
--   SELECT COUNT(id)
--   FROM "product"
-- ),
-- list_products AS (
SELECT * ,
COUNT(*) OVER() AS total_count
FROM "product"
ORDER BY id
LIMIT $1
OFFSET $2;
-- )
-- SELECT *
-- FROM list_products, total_records;

-- name: UpdateProduct :one
UPDATE "product"
SET
category_id = COALESCE(sqlc.narg(category_id),category_id),
brand_id = COALESCE(sqlc.narg(brand_id),brand_id),
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description),
-- product_image = COALESCE(sqlc.narg(product_image),product_image),
active = COALESCE(sqlc.narg(active),active),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM "product"
WHERE id = $1;

-- name: GetProductsByIDs :many
SELECT * FROM "product"
WHERE id = ANY(sqlc.arg(ids)::bigint[]);