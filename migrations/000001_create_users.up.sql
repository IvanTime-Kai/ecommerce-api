CREATE TYPE user_status AS ENUM ('active', 'suspended', 'banned');

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  full_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE,
  phone VARCHAR(16) UNIQUE,
  status user_status NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    email IS NOT NULL
    OR phone IS NOT NULL
  )
);

CREATE INDEX idx_users_status ON users(status);