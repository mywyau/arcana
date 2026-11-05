CREATE TABLE accounts (
    account_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash TEXT NOT NULL CHECK (password_hash <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE characters
    ADD COLUMN account_id TEXT REFERENCES accounts(account_id) ON DELETE SET NULL;

CREATE UNIQUE INDEX characters_account_id_unique
    ON characters(account_id)
    WHERE account_id IS NOT NULL;
