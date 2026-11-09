package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mywyau/arcana/server/internal/account"
)

var _ account.SessionRepository = (*Store)(nil)

func (store *Store) CreateSession(ctx context.Context, value account.Session) error {
	result, err := store.database.ExecContext(ctx, `
		INSERT INTO account_sessions (token_hash, account_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, value.TokenHash[:], value.AccountID, value.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert account session: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect account session insert: %w", err)
	}
	if rows == 0 {
		return account.ErrSessionExists
	}
	return nil
}

func (store *Store) FindSession(
	ctx context.Context,
	tokenHash [sha256.Size]byte,
) (account.Session, bool, error) {
	var value account.Session
	value.TokenHash = tokenHash
	err := store.database.QueryRowContext(ctx, `
		SELECT account_sessions.account_id, characters.player_id, accounts.email,
		       account_sessions.expires_at
		FROM account_sessions
		JOIN accounts ON accounts.account_id = account_sessions.account_id
		JOIN characters ON characters.account_id = account_sessions.account_id
		WHERE account_sessions.token_hash = $1
		  AND account_sessions.expires_at > now()`, tokenHash[:],
	).Scan(&value.AccountID, &value.PlayerID, &value.Email, &value.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return account.Session{}, false, nil
	}
	if err != nil {
		return account.Session{}, false, fmt.Errorf("query account session: %w", err)
	}
	return value, true, nil
}

func (store *Store) DeleteSession(
	ctx context.Context,
	tokenHash [sha256.Size]byte,
) error {
	if _, err := store.database.ExecContext(
		ctx,
		"DELETE FROM account_sessions WHERE token_hash = $1",
		tokenHash[:],
	); err != nil {
		return fmt.Errorf("delete account session: %w", err)
	}
	return nil
}
