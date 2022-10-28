-- name: CreateUser :one
INSERT INTO "user" (
  username,
  email,
  password,
  telephone,
  default_payment
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
WHERE email = $1 LIMIT 1;

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
telephone = COALESCE(sqlc.narg(telephone),telephone),
default_payment = COALESCE(sqlc.narg(default_payment),default_payment)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;