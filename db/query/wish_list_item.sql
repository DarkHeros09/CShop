-- name: CreateWishListItem :one
INSERT INTO "wish_list_item" (
  wish_list_id,
  product_item_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetWishListItem :one
SELECT * FROM "wish_list_item"
WHERE id = $1 LIMIT 1;

-- name: ListWishListItems :many
SELECT * FROM "wish_list_item"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListWishListItemsByCartID :many
SELECT * FROM "wish_list_item"
WHERE wish_list_id = $1
ORDER BY id;

-- name: UpdateWishListItem :one
UPDATE "wish_list_item"
SET 
wish_list_id = COALESCE(sqlc.narg(wish_list_id),wish_list_id),
product_item_id = COALESCE(sqlc.narg(product_item_id),product_item_id)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteWishListItem :exec
DELETE FROM "wish_list_item"
WHERE id = $1;