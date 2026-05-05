CREATE TYPE order_status AS ENUM (
  'pending',
  'confirmed',
  'shipping',
  'delivered',
  'cancelled'
);

CREATE TABLE IF NOT EXISTS orders (
  id UUID NOT NULL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  shop_id UUID NOT NULL REFERENCES shops(id),
  status order_status NOT NULL DEFAULT 'pending',
  total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
  shipping_full_name VARCHAR(255) NOT NULL,
  shipping_phone VARCHAR(20) NOT NULL,
  shipping_province VARCHAR(100) NOT NULL,
  shipping_district VARCHAR(100) NOT NULL,
  shipping_ward VARCHAR(100) NOT NULL,
  shipping_street TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
  id UUID NOT NULL PRIMARY KEY,
  order_id UUID NOT NULL REFERENCES orders(id),
  product_id UUID NOT NULL REFERENCES products(id),
  product_name VARCHAR(255) NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 0,
  price NUMERIC(12, 2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);