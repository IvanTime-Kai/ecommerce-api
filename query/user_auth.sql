-- name: CreateUserAuth :one
INSERT INTO
  user_auth (user_id, provider, password_hash)
VALUES
  ($1, $2, $3) RETURNING *;