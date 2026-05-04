-- name: CreateShop :one
INSERT INTO
  shops (id, owner_id, name, slug, description, logo_url)
VALUES
  ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GeShopByID :one
SELECT
  *
FROM
  shops
WHERE
  id = $1;

-- name: GetShopByOwnerID :one
SELECT
  *
FROM
  shops
WHERE
  owner_id = $1;

-- name: GetShopBySlug :one 
SELECT
  *
FROM
  shops
WHERE
  slug = $1;

-- name: UpdateShop :one
UPDATE
  shops
SET
  name = $2,
  description = $3,
  logo_url = $4,
  updated_at = NOW()
WHERE
  id = $1 RETURNING *;