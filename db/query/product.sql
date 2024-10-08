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

-- name: AdminUpdateProduct :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
UPDATE "product"
SET
category_id = COALESCE(sqlc.narg(category_id),category_id),
brand_id = COALESCE(sqlc.narg(brand_id),brand_id),
name = COALESCE(sqlc.narg(name),name),
description = COALESCE(sqlc.narg(description),description),
-- product_image = COALESCE(sqlc.narg(product_image),product_image),
active = COALESCE(sqlc.narg(active),active),
updated_at = now()
WHERE "product".id = sqlc.arg(id)
AND (SELECT is_admin FROM t1) = 1
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM "product"
WHERE id = $1;

-- name: AdminDeleteProduct :exec
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
DELETE FROM "product" AS p
WHERE p.id = $1
AND (SELECT is_admin FROM t1) = 1;

-- name: GetProductsByIDs :many
SELECT * FROM "product"
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: ListProductsV2 :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
ORDER BY id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: ListProductsNextPage :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
WHERE
 p.id < sqlc.arg(id) 
ORDER BY id DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: SearchProducts :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p
WHERE 
p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)::VARCHAR) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
ORDER BY 
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;

-- name: SearchProductsNextPage :many
WITH t1 AS(
SELECT 
 p.id, p.name, p.description, p.category_id, p.brand_id, p.active, p.created_at, p.updated_at
FROM "product" AS p

WHERE 
p.id < sqlc.arg(product_id) AND
p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)::VARCHAR) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*')::VARCHAR)
    ELSE to_tsquery(sqlc.arg(query)::VARCHAR)
END
ORDER BY 
p.id DESC,
ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)::VARCHAR) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*')::VARCHAR)
    ELSE to_tsquery(sqlc.arg(query)::VARCHAR)
END
) DESC
LIMIT $1 +1
)

SELECT *,COUNT(*) OVER()>10 AS next_available FROM t1 
LIMIT $1;