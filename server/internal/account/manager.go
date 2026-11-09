package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	accountIDBytes    = 16
	sessionTokenBytes = 32
	sessionLifetime   = 30 * 24 * time.Hour
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidSession     = errors.New("invalid or expired account session")
)

type Authentication struct {
	PlayerID     string
	Email        string
	SessionToken string
}

type Manager struct {
	accounts Repository
	sessions SessionRepository
	hasher   PasswordHasher
}

func NewManager(accounts Repository, sessions SessionRepository) (*Manager, error) {
	if accounts == nil {
		return nil, fmt.Errorf("account repository is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("account session repository is required")
	}
	return &Manager{accounts: accounts, sessions: sessions, hasher: NewPasswordHasher()}, nil
}

// Register links the currently authenticated guest character to a recoverable
// account and issues its first session.
func (manager *Manager) Register(
	ctx context.Context,
	playerID string,
	email string,
	password string,
) (Authentication, error) {
	if playerID == "" {
		return Authentication{}, fmt.Errorf("player ID is required")
	}
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return Authentication{}, err
	}
	passwordHash, err := manager.hasher.Hash(password)
	if err != nil {
		return Authentication{}, err
	}
	accountID, err := randomID("account-", accountIDBytes)
	if err != nil {
		return Authentication{}, err
	}
	if err := manager.accounts.CreateAccount(ctx, Account{
		ID: accountID, Email: normalizedEmail, PasswordHash: passwordHash, PlayerID: playerID,
	}); err != nil {
		return Authentication{}, err
	}
	return manager.issueSession(ctx, accountID, playerID, normalizedEmail)
}

func (manager *Manager) Login(
	ctx context.Context,
	email string,
	password string,
) (Authentication, error) {
	value, found, err := manager.accounts.FindAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrInvalidEmail) {
			return Authentication{}, ErrInvalidCredentials
		}
		return Authentication{}, err
	}
	if !found {
		return Authentication{}, ErrInvalidCredentials
	}
	matched, err := manager.hasher.Verify(value.PasswordHash, password)
	if err != nil {
		return Authentication{}, fmt.Errorf("verify password: %w", err)
	}
	if !matched {
		return Authentication{}, ErrInvalidCredentials
	}
	return manager.issueSession(ctx, value.ID, value.PlayerID, value.Email)
}

func (manager *Manager) AuthenticateSession(
	ctx context.Context,
	token string,
) (Authentication, error) {
	rawToken, err := decodeSessionToken(token)
	if err != nil {
		return Authentication{}, err
	}
	value, found, err := manager.sessions.FindSession(ctx, sha256.Sum256(rawToken))
	if err != nil {
		return Authentication{}, fmt.Errorf("find account session: %w", err)
	}
	if !found {
		return Authentication{}, ErrInvalidSession
	}
	return Authentication{PlayerID: value.PlayerID, Email: value.Email}, nil
}

// Logout revokes one device session. Deleting an already-revoked valid token
// is harmless, which makes retries safe.
func (manager *Manager) Logout(ctx context.Context, token string) error {
	rawToken, err := decodeSessionToken(token)
	if err != nil {
		return err
	}
	if err := manager.sessions.DeleteSession(ctx, sha256.Sum256(rawToken)); err != nil {
		return fmt.Errorf("delete account session: %w", err)
	}
	return nil
}

func decodeSessionToken(token string) ([]byte, error) {
	rawToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(rawToken) != sessionTokenBytes {
		return nil, ErrInvalidSession
	}
	return rawToken, nil
}

func (manager *Manager) issueSession(
	ctx context.Context,
	accountID string,
	playerID string,
	email string,
) (Authentication, error) {
	for {
		rawToken := make([]byte, sessionTokenBytes)
		if _, err := rand.Read(rawToken); err != nil {
			return Authentication{}, fmt.Errorf("generate account session: %w", err)
		}
		err := manager.sessions.CreateSession(ctx, Session{
			TokenHash: sha256.Sum256(rawToken),
			AccountID: accountID,
			PlayerID:  playerID,
			Email:     email,
			ExpiresAt: time.Now().Add(sessionLifetime),
		})
		if errors.Is(err, ErrSessionExists) {
			continue
		}
		if err != nil {
			return Authentication{}, fmt.Errorf("create account session: %w", err)
		}
		return Authentication{
			PlayerID:     playerID,
			Email:        email,
			SessionToken: base64.RawURLEncoding.EncodeToString(rawToken),
		}, nil
	}
}

func randomID(prefix string, byteCount int) (string, error) {
	rawID := make([]byte, byteCount)
	if _, err := rand.Read(rawID); err != nil {
		return "", fmt.Errorf("generate account ID: %w", err)
	}
	return prefix + hex.EncodeToString(rawID), nil
}
