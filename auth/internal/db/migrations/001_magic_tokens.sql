CREATE TABLE magic_tokens (
    value       TEXT PRIMARY KEY,
    email       TEXT NOT NULL,
    created_at  INTEGER NOT NULL,
    expires_at  INTEGER NOT NULL
);

CREATE INDEX idx_magic_tokens_expires ON magic_tokens(expires_at);
