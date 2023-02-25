-- name: CreateWishList :one
INSERT INTO "wish_list" (
  user_id
) VALUES (
  $1
)
RETURNING *;

-- name: GetWishList :one
SELECT * FROM "wish_list"
WHERE id = $1 LIMIT 1;

-- name: GetWishListByUserID :one
SELECT * FROM "wish_list"
WHERE user_id = $1 LIMIT 1;

-- name: ListWishLists :many
SELECT * FROM "wish_list"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateWishList :one
UPDATE "wish_list"
SET 
user_id = COALESCE(sqlc.narg(user_id),user_id),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteWishList :exec
DELETE FROM "wish_list"
WHERE id = $1;