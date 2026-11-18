package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

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
		UPDATE characters SET account_id = $1, character_name = $3, updated_at = now()
		WHERE player_id = $2 AND account_id IS NULL`, value.ID, value.PlayerID, value.CharacterName)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.ConstraintName == "characters_character_name_lower_unique" {
			return account.ErrCharacterNameExists
		}
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
		SELECT accounts.account_id, accounts.email, characters.character_name,
		       accounts.password_hash, characters.player_id
		FROM accounts
		JOIN characters ON characters.account_id = accounts.account_id
		WHERE accounts.email = $1`, normalizedEmail,
	).Scan(&value.ID, &value.Email, &value.CharacterName, &value.PasswordHash, &value.PlayerID)
	if errors.Is(err, sql.ErrNoRows) {
		return account.Account{}, false, nil
	}
	if err != nil {
		return account.Account{}, false, fmt.Errorf("query account by email: %w", err)
	}
	return value, true, nil
}

func (store *Store) FindAccountByLogin(
	ctx context.Context,
	identifier string,
) (account.Account, bool, error) {
	identifier = strings.TrimSpace(identifier)
	var value account.Account
	var err error
	if normalizedEmail, emailErr := account.NormalizeEmail(identifier); emailErr == nil {
		err = store.database.QueryRowContext(ctx, `
			SELECT accounts.account_id, accounts.email, characters.character_name,
			       accounts.password_hash, characters.player_id
			FROM accounts
			JOIN characters ON characters.account_id = accounts.account_id
			WHERE accounts.email = $1`, normalizedEmail,
		).Scan(&value.ID, &value.Email, &value.CharacterName, &value.PasswordHash, &value.PlayerID)
	} else {
		normalizedName, nameErr := account.NormalizeCharacterName(identifier)
		if nameErr != nil {
			return account.Account{}, false, nameErr
		}
		err = store.database.QueryRowContext(ctx, `
			SELECT accounts.account_id, accounts.email, characters.character_name,
			       accounts.password_hash, characters.player_id
			FROM accounts
			JOIN characters ON characters.account_id = accounts.account_id
			WHERE lower(characters.character_name) = lower($1)`, normalizedName,
		).Scan(&value.ID, &value.Email, &value.CharacterName, &value.PasswordHash, &value.PlayerID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return account.Account{}, false, nil
	}
	if err != nil {
		return account.Account{}, false, fmt.Errorf("query account by login: %w", err)
	}
	return value, true, nil
}

func (store *Store) DeleteAccount(ctx context.Context, accountID string) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account deletion: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		"UPDATE characters SET character_name = NULL WHERE account_id = $1", accountID,
	); err != nil {
		return fmt.Errorf("clear character name: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM accounts WHERE account_id = $1", accountID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account deletion: %w", err)
	}
	return nil
}
