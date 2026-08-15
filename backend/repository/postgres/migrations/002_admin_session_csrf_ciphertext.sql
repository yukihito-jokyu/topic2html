ALTER TABLE admin_sessions ADD COLUMN csrf_token_ciphertext BYTEA NULL;

UPDATE admin_sessions
SET revoked_at = CURRENT_TIMESTAMP
WHERE csrf_token_ciphertext IS NULL AND revoked_at IS NULL;
