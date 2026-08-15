CREATE TABLE IF NOT EXISTS admin_oauth_transactions (
  id UUID PRIMARY KEY,
  reference_hash BYTEA NOT NULL UNIQUE,
  state_hash BYTEA NOT NULL UNIQUE,
  nonce_hash BYTEA NOT NULL,
  pkce_verifier_ciphertext BYTEA NOT NULL,
  return_path TEXT NOT NULL CHECK (return_path = '/admin'),
  created_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ NULL,
  invalidated_at TIMESTAMPTZ NULL
);

CREATE INDEX admin_oauth_transactions_expires_at_idx ON admin_oauth_transactions (expires_at);

CREATE TABLE IF NOT EXISTS admin_sessions (
  id UUID PRIMARY KEY,
  reference_hash BYTEA NOT NULL UNIQUE,
  authorized_email TEXT NOT NULL,
  csrf_token_hash BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  last_mutation_at TIMESTAMPTZ NOT NULL,
  absolute_expires_at TIMESTAMPTZ NOT NULL,
  idle_expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL,
  CHECK (idle_expires_at <= absolute_expires_at)
);

CREATE INDEX admin_sessions_absolute_expires_at_idx ON admin_sessions (absolute_expires_at);
CREATE INDEX admin_sessions_revoked_at_idx ON admin_sessions (revoked_at);
