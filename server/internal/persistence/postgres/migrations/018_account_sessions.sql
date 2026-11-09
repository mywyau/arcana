CREATE TABLE account_sessions (
    token_hash BYTEA PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    account_id TEXT NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL CHECK (expires_at > created_at)
);

CREATE INDEX account_sessions_account_id_idx
    ON account_sessions(account_id);

CREATE INDEX account_sessions_expires_at_idx
    ON account_sessions(expires_at);
