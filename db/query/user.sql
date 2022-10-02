-- name: CreateUser :one
INSERT INTO "user" (
  email,
  password,
  telephone
) VALUES (
  $1, $2, $3
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
email = COALESCE(sqlc.narg(email),email),
password = COALESCE(sqlc.narg(password),password),
telephone = COALESCE(sqlc.narg(telephone),telephone)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;