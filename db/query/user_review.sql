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
WHERE id = $1 
AND user_id = $2
LIMIT 1;

-- name: ListUserReviews :many
SELECT * FROM "user_review"
WHERE user_id = $3
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUserReview :one
UPDATE "user_review"
SET 
ordered_product_id = COALESCE(sqlc.narg(ordered_product_id),ordered_product_id),
rating_value = COALESCE(sqlc.narg(rating_value),rating_value),
updated_at = now()
WHERE id = sqlc.arg(id)
AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteUserReview :one
DELETE FROM "user_review"
WHERE id = $1
And user_id =$2
RETURNING *;