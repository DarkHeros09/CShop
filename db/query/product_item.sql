-- name: CreateProductItem :one
INSERT INTO "product_item" (
  product_id,
  product_sku,
  qty_in_stock,
  product_image,
  price,
  active
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetProductItem :one
SELECT * FROM "product_item"
WHERE id = $1 LIMIT 1;

-- name: GetProductItemForUpdate :one
SELECT * FROM "product_item"
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListProductItemsByIDs :many
SELECT pi.id, p.name, pi.product_id, 
pi.product_image, pi.price, pi.active
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id 
WHERE pi.id = ANY(sqlc.arg(products_ids)::bigint[]);

-- name: ListProductItems :many
SELECT pi.*, p.name, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id 
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListProductItemsV2 :many
SELECT pi.*, p.name, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id 
ORDER BY pi.id DESC
LIMIT $1;

-- name: ListProductItemsNextPage :many
WITH t1 AS (
SELECT COUNT(*) OVER() AS total_count
FROM "product_item" AS p
LIMIT 1
)
SELECT pi.*, p.name, (SELECT total_count FROM t1)
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id 
WHERE pi.id < $2
ORDER BY pi.id DESC
LIMIT $1;

-- name: SearchProductItems :many
SELECT pi.*, p.name, COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
WHERE p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
ORDER BY pi.id DESC, ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1;

-- name: SearchProductItemsNextPage :many
-- WITH t1 AS (
-- SELECT COUNT(*) OVER() AS total_count
-- FROM "product_item" AS p
-- LIMIT 1
-- )
SELECT pi.*, p.name,  COUNT(*) OVER() AS total_count
FROM "product_item" AS pi
LEFT JOIN "product" AS p ON p.id = pi.product_id
WHERE p.search @@ 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
AND pi.id < $2
ORDER BY pi.id DESC, ts_rank(p.search, 
CASE
    WHEN char_length(sqlc.arg(query)) > 0 THEN to_tsquery(concat(sqlc.arg(query), ':*'))
    ELSE to_tsquery(sqlc.arg(query))
END
) DESC
LIMIT $1;

-- name: UpdateProductItem :one
UPDATE "product_item"
SET
product_sku = COALESCE(sqlc.narg(product_sku),product_sku),
qty_in_stock = COALESCE(sqlc.narg(qty_in_stock),qty_in_stock),
product_image = COALESCE(sqlc.narg(product_image),product_image),
price = COALESCE(sqlc.narg(price),price),
active = COALESCE(sqlc.narg(active),active),
updated_at = now()
WHERE id = sqlc.arg(id)
AND product_id = sqlc.arg(product_id)
RETURNING *;

-- name: DeleteProductItem :exec
DELETE FROM "product_item"
WHERE id = $1;