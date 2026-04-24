CREATE TYPE provider_type AS ENUM('local', 'google');

CREATE TABLE IF NOT EXISTS user_auth (
  user_id UUID NOT NULL REFERENCES users(id),
  provider provider_type DEFAULT 'local',
  provider_id TEXT,
  password_hash TEXT,
  is_verified BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, provider),
  CHECK (
    provider != 'local'
    OR password_hash IS NOT NULL
  )
);