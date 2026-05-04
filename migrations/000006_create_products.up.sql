CREATE TYPE product_status AS ENUM ('active', 'inactive', 'deleted');

CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY,
  shop_id UUID NOT NULL REFERENCES shops(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  status product_status NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);