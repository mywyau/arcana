package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const maximumEmailBytes = 254

var (
	ErrAccountExists        = errors.New("account already exists")
	ErrCharacterUnavailable = errors.New("character is unavailable")
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
}

// NormalizeEmail defines Arcana's account identity rules. Email addresses are
// deliberately case-insensitive for login and uniqueness.
func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || len(normalized) > maximumEmailBytes || strings.ContainsAny(normalized, " \t\r\n") {
		return "", fmt.Errorf("email address is invalid")
	}
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", fmt.Errorf("email address is invalid")
	}
	return normalized, nil
}

type MemoryRepository struct {
	mu              sync.RWMutex
	accountsByID    map[string]Account
	accountByEmail  map[string]string
	accountByPlayer map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		accountsByID:    make(map[string]Account),
		accountByEmail:  make(map[string]string),
		accountByPlayer: make(map[string]string),
	}
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
