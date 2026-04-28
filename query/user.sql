-- name: CreateUser :one
INSERT INTO
  users (id, full_name, email, phone)
VALUES
  ($1, $2, $3, $4) RETURNING *;

-- name: CheckUserEmailExists :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      users
    WHERE
      email = $1
  );

-- name: CheckUserPhoneExists :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      users
    WHERE
      phone = $1
  );

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;