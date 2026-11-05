package account

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryRepositoryNormalizesAndFindsAccount(t *testing.T) {
	repository := NewMemoryRepository()
	want := Account{
		ID: "account-one", Email: "Player@Example.COM ",
		PasswordHash: "encoded-hash", PlayerID: "player-one",
	}
	if err := repository.CreateAccount(context.Background(), want); err != nil {
		t.Fatalf("CreateAccount() returned an unexpected error: %v", err)
	}
	got, found, err := repository.FindAccountByEmail(context.Background(), " PLAYER@example.com")
	if err != nil || !found {
		t.Fatalf("FindAccountByEmail() = found %t, error %v", found, err)
	}
	if got.Email != "player@example.com" || got.ID != want.ID || got.PlayerID != want.PlayerID {
		t.Fatalf("found account = %+v", got)
	}
}

func TestMemoryRepositoryRejectsDuplicateEmailAndCharacter(t *testing.T) {
	repository := NewMemoryRepository()
	first := Account{ID: "one", Email: "one@example.com", PasswordHash: "hash", PlayerID: "player-one"}
	if err := repository.CreateAccount(context.Background(), first); err != nil {
		t.Fatalf("create first account: %v", err)
	}
	duplicateEmail := Account{ID: "two", Email: "ONE@example.com", PasswordHash: "hash", PlayerID: "player-two"}
	if err := repository.CreateAccount(context.Background(), duplicateEmail); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("duplicate email error = %v, want ErrAccountExists", err)
	}
	duplicateCharacter := Account{ID: "three", Email: "three@example.com", PasswordHash: "hash", PlayerID: "player-one"}
	if err := repository.CreateAccount(context.Background(), duplicateCharacter); !errors.Is(err, ErrCharacterUnavailable) {
		t.Fatalf("duplicate character error = %v, want ErrCharacterUnavailable", err)
	}
}

func TestNormalizeEmailRejectsInvalidAddress(t *testing.T) {
	for _, email := range []string{"", "missing-at.example.com", "@example.com", "player@localhost", "two@@example.com"} {
		if _, err := NormalizeEmail(email); err == nil {
			t.Errorf("NormalizeEmail(%q) succeeded", email)
		}
	}
}
