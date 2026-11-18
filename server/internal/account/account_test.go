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
		CharacterName: "Arcane Hero", PasswordHash: "encoded-hash", PlayerID: "player-one",
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
	first := Account{ID: "one", Email: "one@example.com", CharacterName: "First Hero", PasswordHash: "hash", PlayerID: "player-one"}
	if err := repository.CreateAccount(context.Background(), first); err != nil {
		t.Fatalf("create first account: %v", err)
	}
	duplicateEmail := Account{ID: "two", Email: "ONE@example.com", CharacterName: "Second Hero", PasswordHash: "hash", PlayerID: "player-two"}
	if err := repository.CreateAccount(context.Background(), duplicateEmail); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("duplicate email error = %v, want ErrAccountExists", err)
	}
	duplicateCharacter := Account{ID: "three", Email: "three@example.com", CharacterName: "Third Hero", PasswordHash: "hash", PlayerID: "player-one"}
	if err := repository.CreateAccount(context.Background(), duplicateCharacter); !errors.Is(err, ErrCharacterUnavailable) {
		t.Fatalf("duplicate character error = %v, want ErrCharacterUnavailable", err)
	}
}

func TestMemoryRepositoryFindsCharacterNameWithoutCaseSensitivity(t *testing.T) {
	repository := NewMemoryRepository()
	want := Account{
		ID: "account-one", Email: "player@example.com", CharacterName: "Arcane Hero",
		PasswordHash: "encoded-hash", PlayerID: "player-one",
	}
	if err := repository.CreateAccount(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	got, found, err := repository.FindAccountByLogin(context.Background(), "arcane hero")
	if err != nil || !found || got.CharacterName != want.CharacterName {
		t.Fatalf("FindAccountByLogin() = %+v, found %t, error %v", got, found, err)
	}
	duplicateName := Account{
		ID: "account-two", Email: "other@example.com", CharacterName: "ARCANE HERO",
		PasswordHash: "encoded-hash", PlayerID: "player-two",
	}
	if err := repository.CreateAccount(context.Background(), duplicateName); !errors.Is(err, ErrCharacterNameExists) {
		t.Fatalf("duplicate character name error = %v, want ErrCharacterNameExists", err)
	}
}

func TestNormalizeCharacterNameRejectsInvalidNames(t *testing.T) {
	for _, name := range []string{"", "ab", "-hero", "hero_", "hero!", "this-name-is-far-too-long"} {
		if _, err := NormalizeCharacterName(name); err == nil {
			t.Errorf("NormalizeCharacterName(%q) succeeded", name)
		}
	}
}

func TestNormalizeEmailRejectsInvalidAddress(t *testing.T) {
	for _, email := range []string{"", "missing-at.example.com", "@example.com", "player@localhost", "two@@example.com"} {
		if _, err := NormalizeEmail(email); err == nil {
			t.Errorf("NormalizeEmail(%q) succeeded", email)
		}
	}
}
