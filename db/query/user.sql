-- name: CreateUser :one
INSERT INTO "user" (
  username,
  email,
  password,
  -- telephone,
  is_blocked,
  default_payment
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CreateUserWithCartAndWishList :one
WITH t1 AS(
INSERT INTO "user" (
  username,
  email,
  password,
  -- telephone,
  is_blocked,
  default_payment
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *
),
t2 AS(
  INSERT INTO "shopping_cart" (
  user_id
) VALUES ((Select id from t1))
  RETURNING id
),
t3 AS(
  INSERT INTO "wish_list" (
    user_id
) VALUES ((Select id from t1))
  RETURNING id
)

SELECT t1.*, t2.id AS shopping_cart_id, t3.id AS wish_list_id FROM t1, t2, t3;

-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
-- SELECT * FROM "user"
-- WHERE email = $1 LIMIT 1;
SELECT u.*, sc.id AS shop_cart_id, wl.id AS wish_list_id FROM "user" AS u
LEFT JOIN shopping_cart AS sc ON sc.user_id = u.id
LEFT JOIN wish_list AS wl ON wl.user_id = u.id
WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM "user"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE "user"
SET 
username = COALESCE(sqlc.narg(username),username),
email = COALESCE(sqlc.narg(email),email),
password = COALESCE(sqlc.narg(password),password),
-- telephone = COALESCE(sqlc.narg(telephone),telephone),
default_payment = COALESCE(sqlc.narg(default_payment),default_payment),
default_address_id = COALESCE(sqlc.narg(default_address_id),default_address_id),
updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE "user"
SET 
password = sqlc.arg(newPassword),
updated_at = now()
WHERE id = sqlc.arg(id)
AND is_email_verified = TRUE
AND is_blocked = FALSE
AND password = sqlc.arg(oldPassword)
AND password != sqlc.arg(newPassword)
RETURNING *;

-- name: UpdateUserEmailisVerifiedForTest :exec
UPDATE "user"
SET is_email_verified = TRUE
WHERE id = sqlc.arg(id);

-- name: DeleteUser :one
DELETE FROM "user"
WHERE id = $1
RETURNING *;

-- name: DeleteUserByEmailNotVerified :exec
DELETE FROM "user"
WHERE email = $1
AND is_email_verified = false;

-- name: GetActiveUsersCount :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(id) FROM "user"
WHERE EXISTS(SELECT is_admin FROM t1)
AND is_blocked = false;

-- name: GetTotalUsersCount :one
With t1 AS (
SELECT 1 AS is_admin
    FROM "admin"
    WHERE "admin".id = sqlc.arg(admin_id)
    AND active = TRUE
    )
SELECT COUNT(id) FROM "user"
WHERE EXISTS(SELECT is_admin FROM t1);