-- 1. Thêm category_id vào products
ALTER TABLE products
  ADD COLUMN category_id UUID REFERENCES categories(id);

-- 2. Thêm cột search_vector để lưu full-text index
ALTER TABLE products
  ADD COLUMN search_vector tsvector;

-- 3. Index GIN cho search_vector (tại sao GIN? vì tsvector là composite value)
CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- 4. Index thường cho filter
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_price ON products(price);

-- 5. Trigger function: tự update search_vector mỗi khi insert/update
CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('english',
    coalesce(NEW.name, '') || ' ' || coalesce(NEW.description, '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 6. Gắn trigger vào bảng products
CREATE TRIGGER trg_product_search_vector
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_product_search_vector();

-- 7. Populate search_vector cho các product đã có
UPDATE products
SET search_vector = to_tsvector('english',
  coalesce(name, '') || ' ' || coalesce(description, '')
);
