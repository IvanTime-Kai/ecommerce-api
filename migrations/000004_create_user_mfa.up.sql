CREATE TYPE method_type AS ENUM('totp', 'sms', 'email');

CREATE TABLE IF NOT EXISTS user_mfa (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  method method_type NOT NULL DEFAULT 'totp',
  secret_key TEXT NOT NULL,
  backup_codes JSONB NOT NULL DEFAULT '[]',
  is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, method)
);