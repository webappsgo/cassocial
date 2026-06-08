-- Migration 005: token_hash column on email_verification_tokens
-- Migration 001 was updated to create this table with the correct schema.
-- For databases created before that update (token TEXT PRIMARY KEY, no token_hash),
-- this migration adds the token_hash column and index.
-- On fresh databases the ALTER TABLE fails silently (column already exists from migration 001).

ALTER TABLE email_verification_tokens ADD COLUMN token_hash TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_evtoken_hash
    ON email_verification_tokens (token_hash);
