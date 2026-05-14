CREATE TABLE IF NOT EXISTS revenue_summary (
  shop_id UUID NOT NULL REFERENCES shops(id),
  date DATE NOT NULL,
  total NUMERIC(15, 2) NOT NULL DEFAULT 0,
  order_count INT NOT NULL DEFAULT 0,
  PRIMARY KEY (shop_id, date)
);