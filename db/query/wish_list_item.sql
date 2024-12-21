-- name: CreateWishListItem :one
INSERT INTO "wish_list_item" (
  wish_list_id,
  product_item_id,
  size_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetWishListItem :one
SELECT * FROM "wish_list_item"
WHERE id = $1 LIMIT 1;

-- name: GetWishListItemByUserIDCartID :one
SELECT wli.*
FROM "wish_list_item" AS wli
LEFT JOIN "wish_list" AS wl ON wl.id = wli.wish_list_id
WHERE wl.user_id = $1
AND wli.id = $2
AND wli.wish_list_id = $3
LIMIT 1;

-- name: ListWishListItems :many
SELECT * FROM "wish_list_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListWishListItemsByUserID :many
SELECT wl.user_id, wli.*, ps.qty AS size_qty, ps.size_value
FROM "wish_list" AS wl
LEFT JOIN "wish_list_item" AS wli ON wli.wish_list_id = wl.id
JOIN "product_size" AS ps ON wli.size_id = ps.id
WHERE wl.user_id = $1;

-- name: ListWishListItemsByCartID :many
SELECT * FROM "wish_list_item"
WHERE wish_list_id = $1
ORDER BY id;

-- name: UpdateWishListItem :one
-- WITH t1 AS (
--   SELECT user_id FROM "wish_list" AS wl
--   WHERE wl.id = sqlc.arg(wish_list_id)
-- )

UPDATE "wish_list_item" AS wli
SET 
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id),
size_id = COALESCE(sqlc.narg(size_id),size_id),
updated_at = now()
WHERE wli.id = sqlc.arg(id)
AND wli.wish_list_id = sqlc.arg(wish_list_id)
RETURNING *;

-- name: DeleteWishListItem :exec
-- WITH t1 AS (
--   SELECT id FROM "wish_list" AS wl
--   WHERE wl.user_id = sqlc.arg(user_id)
-- )
DELETE FROM "wish_list_item" AS wli
WHERE wli.id = sqlc.arg(id)
AND wli.wish_list_id = sqlc.arg(wish_list_id);

-- name: DeleteWishListItemAll :many
-- WITH t1 AS(
--   SELECT id FROM "wish_list" WHERE user_id = $1
-- )
DELETE FROM "wish_list_item"
WHERE wish_list_id = sqlc.arg(wish_list_id)
RETURNING *;