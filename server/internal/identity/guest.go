package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

const (
	guestTokenBytes = 32
	guestIDBytes    = 16
)

var (
	ErrInvalidGuestToken = errors.New("invalid guest token")
	ErrIdentityExists    = errors.New("guest identity already exists")
)

type GuestRepository interface {
	FindPlayerID(context.Context, [sha256.Size]byte) (string, bool, error)
	Create(context.Context, string, [sha256.Size]byte) error
}

type GuestIdentity struct {
	PlayerID string
	Token    string
}

type GuestManager struct {
	repository GuestRepository
}

func NewGuestManager(repository GuestRepository) (*GuestManager, error) {
	if repository == nil {
		return nil, fmt.Errorf("guest repository is required")
	}
	return &GuestManager{repository: repository}, nil
}

func (manager *GuestManager) Authenticate(ctx context.Context, token string) (GuestIdentity, error) {
	if token != "" {
		rawToken, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil || len(rawToken) != guestTokenBytes {
			return GuestIdentity{}, ErrInvalidGuestToken
		}
		playerID, found, err := manager.repository.FindPlayerID(ctx, sha256.Sum256(rawToken))
		if err != nil {
			return GuestIdentity{}, fmt.Errorf("find guest identity: %w", err)
		}
		if !found {
			return GuestIdentity{}, ErrInvalidGuestToken
		}
		return GuestIdentity{PlayerID: playerID}, nil
	}

	for {
		rawToken := make([]byte, guestTokenBytes)
		if _, err := rand.Read(rawToken); err != nil {
			return GuestIdentity{}, fmt.Errorf("generate guest token: %w", err)
		}
		rawID := make([]byte, guestIDBytes)
		if _, err := rand.Read(rawID); err != nil {
			return GuestIdentity{}, fmt.Errorf("generate guest ID: %w", err)
		}
		playerID := "guest-" + hex.EncodeToString(rawID)
		err := manager.repository.Create(ctx, playerID, sha256.Sum256(rawToken))
		if errors.Is(err, ErrIdentityExists) {
			continue
		}
		if err != nil {
			return GuestIdentity{}, fmt.Errorf("create guest identity: %w", err)
		}
		return GuestIdentity{
			PlayerID: playerID,
			Token:    base64.RawURLEncoding.EncodeToString(rawToken),
		}, nil
	}
}

type MemoryGuestRepository struct {
	mu          sync.RWMutex
	playerIDs   map[string]struct{}
	tokenHashes map[[sha256.Size]byte]string
}

func NewMemoryGuestRepository() *MemoryGuestRepository {
	return &MemoryGuestRepository{
		playerIDs:   make(map[string]struct{}),
		tokenHashes: make(map[[sha256.Size]byte]string),
	}
}

func (repository *MemoryGuestRepository) FindPlayerID(
	_ context.Context,
	tokenHash [sha256.Size]byte,
) (string, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	playerID, found := repository.tokenHashes[tokenHash]
	return playerID, found, nil
}

func (repository *MemoryGuestRepository) Create(
	_ context.Context,
	playerID string,
	tokenHash [sha256.Size]byte,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, exists := repository.playerIDs[playerID]; exists {
		return ErrIdentityExists
	}
	if _, exists := repository.tokenHashes[tokenHash]; exists {
		return ErrIdentityExists
	}
	repository.playerIDs[playerID] = struct{}{}
	repository.tokenHashes[tokenHash] = playerID
	return nil
}
