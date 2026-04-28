-- name: CheckUserEnabledMFA :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      user_mfa
    WHERE
      user_id = $1
      AND is_enabled = true
  );

-- name: GetUserMFAByUserID :one
SELECT * FROM user_mfa WHERE user_id = $1 AND method ='totp';

-- name: GetUserTOTPSecret :one
SELECT secret_key FROM user_mfa WHERE user_id = $1 AND method = 'totp' AND is_enabled = true;

-- name: CreateUserMFA :one
INSERT INTO user_mfa (id, user_id, method, secret_key)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: EnableUserMFA :exec
UPDATE user_mfa SET is_enabled = true WHERE user_id = $1 AND method = 'totp';