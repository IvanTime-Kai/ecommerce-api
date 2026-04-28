-- name: CreateUserAuth :one
INSERT INTO
  user_auth (user_id, provider, password_hash)
VALUES
  ($1, $2, $3) RETURNING *;

-- name: GetUserPasswordHash :one 
SELECT
  password_hash
FROM
  user_auth
WHERE
  user_id = $1 AND provider = $2;