-- name: CreateVerifyEmail :one
INSERT INTO "verify_email" (
    user_id,
    -- email,
    secret_code
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetVerifyEmail :one
SELECT * FROM "verify_email"
WHERE id = $1 LIMIT 1;

-- name: GetVerifyEmailByEmail :one
SELECT u.email, u.username, u.is_blocked, u.is_email_verified, 
ve.* FROM "verify_email" AS ve
JOIN "user" AS u ON u.id = ve.user_id
WHERE u.email = $1
ORDER BY ve.created_at DESC
LIMIT 1;

-- name: UpdateVerifyEmail :one
with t1 AS (
SELECT id FROM "user" AS u
WHERE u.email = sqlc.arg(email)
AND u.is_email_verified = FALSE
),
t2 AS (
UPDATE "verify_email" 
SET is_used = TRUE
WHERE secret_code = sqlc.arg(secret_code)
AND user_id = (SELECT id FROM t1)
AND expired_at > NOW()
RETURNING *
),
t3 AS (
UPDATE "user"
SET is_email_verified = TRUE
WHERE id = (SELECT user_id FROM t2)
RETURNING id, username, email, is_blocked, is_email_verified, default_payment, created_at, updated_at
),
t4 AS(
  INSERT INTO "shopping_cart" (
  user_id
) VALUES ((SELECT id FROM t1))
  RETURNING id
),
t5 AS(
  INSERT INTO "wish_list" (
    user_id
) VALUES ((SELECT id FROM t1))
  RETURNING id
)

SELECT t3.*, t4.id AS shopping_cart_id, t5.id AS wish_list_id FROM t3,t4,t5;