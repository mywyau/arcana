package identity

import (
	"context"
	"errors"
	"testing"
)

func TestGuestManagerCreatesAndResumesIdentity(t *testing.T) {
	manager, err := NewGuestManager(NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	created, err := manager.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("Authenticate(new) returned an unexpected error: %v", err)
	}
	if created.PlayerID == "" || created.Token == "" {
		t.Fatalf("created identity = %+v, want player ID and token", created)
	}
	resumed, err := manager.Authenticate(context.Background(), created.Token)
	if err != nil {
		t.Fatalf("Authenticate(existing) returned an unexpected error: %v", err)
	}
	if resumed.PlayerID != created.PlayerID || resumed.Token != "" {
		t.Fatalf("resumed identity = %+v, want player %q and no replacement token", resumed, created.PlayerID)
	}
}

func TestGuestManagerRejectsUnknownAndMalformedTokens(t *testing.T) {
	manager, _ := NewGuestManager(NewMemoryGuestRepository())
	otherManager, _ := NewGuestManager(NewMemoryGuestRepository())
	other, err := otherManager.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("Authenticate(new) returned an unexpected error: %v", err)
	}
	for _, token := range []string{"not-a-token", other.Token} {
		if _, err := manager.Authenticate(context.Background(), token); !errors.Is(err, ErrInvalidGuestToken) {
			t.Fatalf("Authenticate(%q) error = %v, want invalid guest token", token, err)
		}
	}
}
