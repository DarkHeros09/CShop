-- name: CreateUserReview :one
INSERT INTO "user_review" (
  user_id,
  ordered_product_id,
  rating_value
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserReview :one
SELECT * FROM "user_review"
WHERE id = $1 LIMIT 1;

-- name: ListUserReviews :many
SELECT * FROM "user_review"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUserReview :one
UPDATE "user_review"
SET 
user_id = COALESCE(sqlc.narg(user_id),user_id),
ordered_product_id = COALESCE(sqlc.narg(ordered_product_id),ordered_product_id),
rating_value = COALESCE(sqlc.narg(rating_value),rating_value)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUserReview :exec
DELETE FROM "user_review"
WHERE id = $1;