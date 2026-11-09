package account

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const maximumEmailBytes = 254

var (
	ErrAccountExists        = errors.New("account already exists")
	ErrCharacterUnavailable = errors.New("character is unavailable")
	ErrInvalidEmail         = errors.New("email address is invalid")
	ErrSessionExists        = errors.New("account session already exists")
)

// Account is the private authentication record linked to one character.
type Account struct {
	ID           string
	Email        string
	PasswordHash string
	PlayerID     string
}

type Repository interface {
	CreateAccount(context.Context, Account) error
	FindAccountByEmail(context.Context, string) (Account, bool, error)
	DeleteAccount(context.Context, string) error
}

// Session contains only a token hash. Raw session tokens are returned to the
// client once and are never persisted by a repository.
type Session struct {
	TokenHash [sha256.Size]byte
	AccountID string
	PlayerID  string
	Email     string
	ExpiresAt time.Time
}

type SessionRepository interface {
	CreateSession(context.Context, Session) error
	FindSession(context.Context, [sha256.Size]byte) (Session, bool, error)
	DeleteSession(context.Context, [sha256.Size]byte) error
}

// NormalizeEmail defines Arcana's account identity rules. Email addresses are
// deliberately case-insensitive for login and uniqueness.
func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || len(normalized) > maximumEmailBytes || strings.ContainsAny(normalized, " \t\r\n") {
		return "", ErrInvalidEmail
	}
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", ErrInvalidEmail
	}
	return normalized, nil
}

type MemoryRepository struct {
	mu              sync.RWMutex
	accountsByID    map[string]Account
	accountByEmail  map[string]string
	accountByPlayer map[string]string
	sessions        map[[sha256.Size]byte]Session
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		accountsByID:    make(map[string]Account),
		accountByEmail:  make(map[string]string),
		accountByPlayer: make(map[string]string),
		sessions:        make(map[[sha256.Size]byte]Session),
	}
}

func (repository *MemoryRepository) CreateSession(_ context.Context, value Session) error {
	if value.AccountID == "" || value.PlayerID == "" || value.ExpiresAt.IsZero() {
		return fmt.Errorf("account session is incomplete")
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, exists := repository.sessions[value.TokenHash]; exists {
		return ErrSessionExists
	}
	repository.sessions[value.TokenHash] = value
	return nil
}

func (repository *MemoryRepository) FindSession(
	_ context.Context,
	tokenHash [sha256.Size]byte,
) (Session, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	value, found := repository.sessions[tokenHash]
	if !found || !value.ExpiresAt.After(time.Now()) {
		return Session{}, false, nil
	}
	return value, true, nil
}

func (repository *MemoryRepository) DeleteSession(
	_ context.Context,
	tokenHash [sha256.Size]byte,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	delete(repository.sessions, tokenHash)
	return nil
}

func (repository *MemoryRepository) CreateAccount(_ context.Context, value Account) error {
	value, err := NormalizeAccount(value)
	if err != nil {
		return err
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, exists := repository.accountsByID[value.ID]; exists {
		return ErrAccountExists
	}
	if _, exists := repository.accountByEmail[value.Email]; exists {
		return ErrAccountExists
	}
	if _, exists := repository.accountByPlayer[value.PlayerID]; exists {
		return ErrCharacterUnavailable
	}
	repository.accountsByID[value.ID] = value
	repository.accountByEmail[value.Email] = value.ID
	repository.accountByPlayer[value.PlayerID] = value.ID
	return nil
}

func (repository *MemoryRepository) FindAccountByEmail(
	_ context.Context,
	email string,
) (Account, bool, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return Account{}, false, err
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	accountID, found := repository.accountByEmail[normalizedEmail]
	if !found {
		return Account{}, false, nil
	}
	return repository.accountsByID[accountID], true, nil
}

func (repository *MemoryRepository) DeleteAccount(_ context.Context, accountID string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	value, found := repository.accountsByID[accountID]
	if !found {
		return nil
	}
	delete(repository.accountsByID, accountID)
	delete(repository.accountByEmail, value.Email)
	delete(repository.accountByPlayer, value.PlayerID)
	for tokenHash, session := range repository.sessions {
		if session.AccountID == accountID {
			delete(repository.sessions, tokenHash)
		}
	}
	return nil
}

func NormalizeAccount(value Account) (Account, error) {
	if value.ID == "" {
		return Account{}, fmt.Errorf("account ID is required")
	}
	if value.PlayerID == "" {
		return Account{}, fmt.Errorf("player ID is required")
	}
	if value.PasswordHash == "" {
		return Account{}, fmt.Errorf("password hash is required")
	}
	normalizedEmail, err := NormalizeEmail(value.Email)
	if err != nil {
		return Account{}, err
	}
	value.Email = normalizedEmail
	return value, nil
}
