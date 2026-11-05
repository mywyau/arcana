package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mywyau/arcana/server/internal/account"
)

var _ account.Repository = (*Store)(nil)

func (store *Store) CreateAccount(ctx context.Context, value account.Account) error {
	value, err := account.NormalizeAccount(value)
	if err != nil {
		return err
	}
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account creation: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (account_id, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, value.ID, value.Email, value.PasswordHash)
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect account insert: %w", err)
	}
	if rows == 0 {
		return account.ErrAccountExists
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE characters SET account_id = $1, updated_at = now()
		WHERE player_id = $2 AND account_id IS NULL`, value.ID, value.PlayerID)
	if err != nil {
		return fmt.Errorf("link character to account: %w", err)
	}
	rows, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect character link: %w", err)
	}
	if rows == 0 {
		return account.ErrCharacterUnavailable
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account creation: %w", err)
	}
	return nil
}

func (store *Store) FindAccountByEmail(
	ctx context.Context,
	email string,
) (account.Account, bool, error) {
	normalizedEmail, err := account.NormalizeEmail(email)
	if err != nil {
		return account.Account{}, false, err
	}
	var value account.Account
	err = store.database.QueryRowContext(ctx, `
		SELECT accounts.account_id, accounts.email, accounts.password_hash, characters.player_id
		FROM accounts
		JOIN characters ON characters.account_id = accounts.account_id
		WHERE accounts.email = $1`, normalizedEmail,
	).Scan(&value.ID, &value.Email, &value.PasswordHash, &value.PlayerID)
	if errors.Is(err, sql.ErrNoRows) {
		return account.Account{}, false, nil
	}
	if err != nil {
		return account.Account{}, false, fmt.Errorf("query account by email: %w", err)
	}
	return value, true, nil
}
