-- name: CreateProduct :one
INSERT INTO
  products (id, shop_id, name, description, price, stock)
VALUES
  ($1, $2, $3, $4, $5, $6) RETURNING *;

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
  AND status = 'active';

-- name: UpdateProduct :one
UPDATE
  products
SET
  name = $2,
  description = $3,
  status = $4,
  updated_at = NOW()
WHERE
  id = $1 RETURNING *;

-- name: DeleteProduct :exec
UPDATE
  products
SET
  status = 'deleted',
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

-- name: GetProductsForOrder :many
SELECT
  id,
  name,
  price,
  stock,
  shop_id
FROM
  products
WHERE
  id = ANY($1::uuid[])
  AND shop_id = $2
  AND status = 'active'
ORDER BY
  id FOR
UPDATE
;

-- name: DeductProductStock :one
UPDATE
  products
SET
  stock = stock - $2,
  updated_at = NOW()
WHERE
  id = $1
  AND stock >= $2 RETURNING *;

-- name: GetAllProducts :many
SELECT * FROM products;
