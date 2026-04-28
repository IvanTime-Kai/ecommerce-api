-- name: CreateUserSession :one
INSERT INTO
  auth_sessions (
    id,
    user_id,
    ip,
    user_agent,
    refresh_token_hash,
    expires_at,
    is_remember
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM auth_sessions WHERE refresh_token_hash = $1;

-- name: DeleteSessionByRefreshToken :exec
DELETE FROM auth_sessions WHERE refresh_token_hash = $1;