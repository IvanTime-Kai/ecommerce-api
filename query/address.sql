-- name: CreateAddress :one
INSERT INTO
  addresses (
    id,
    user_id,
    full_name,
    phone,
    province,
    district,
    ward,
    street,
    is_default
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: GetAddressByID :one
SELECT
  *
FROM
  addresses
WHERE
  id = $1;

-- name: GetAddressesByUserID :one
SELECT
  *
FROM
  addresses
WHERE
  user_id = $1; 

-- name: SetDefaultAddress :exec
UPDATE
  addresses
SET
  is_default = false
WHERE
  user_id = $1;

-- name: UpdateAddressDefault :exec
UPDATE
  addresses
SET
  is_default = true
WHERE
  id = $1;

-- name: DeleteAddress :exec
DELETE FROM
  addresses
WHERE
  id = $1;