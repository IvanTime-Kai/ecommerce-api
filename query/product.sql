-- name: CreateProduct :one
INSERT INTO
  products (id, shop_id, name, description)
VALUES
  ($1, $2, $3, $4) RETURNING *;

-- name: GetProductByID :one
SELECT
  *
FROM
  products
WHERE
  id = $1;

-- name: GetProductsByShopID :many
SELECT
  *
FROM
  products
WHERE
  shop_id = $1
  AND is_active = true;

-- name: UpdateProduct :one
UPDATE
  products
SET
  name = $2,
  description = $3,
  is_active = $4,
  updated_at = NOW()
WHERE
  id = $1 RETURNING *;

-- name: DeleteProduct :exec
UPDATE
  products
SET
  status = 'deleted',
  is_active = false,
  updated_at = NOW()
WHERE
  id = $1;

-- name: GetProductByIDAndShopOwner :one
SELECT
  p.*
FROM
  products p
  JOIN shops s ON s.id = p.shop_id
WHERE
  p.id = $1
  AND s.owner_id = $2;