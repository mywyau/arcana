package account

import (
	"context"
	"errors"
	"testing"
)

func TestManagerRegistersLogsInAndAuthenticatesSession(t *testing.T) {
	repository := NewMemoryRepository()
	manager, err := NewManager(repository, repository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	registered, err := manager.Register(
		context.Background(), "player-one", " Player@Example.com ", "strong-password",
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.PlayerID != "player-one" || registered.Email != "player@example.com" || registered.SessionToken == "" {
		t.Fatalf("registered authentication = %+v", registered)
	}

	fromSession, err := manager.AuthenticateSession(context.Background(), registered.SessionToken)
	if err != nil || fromSession.PlayerID != "player-one" {
		t.Fatalf("AuthenticateSession() = %+v, error %v", fromSession, err)
	}
	loggedIn, err := manager.Login(context.Background(), "PLAYER@example.com", "strong-password")
	if err != nil || loggedIn.PlayerID != "player-one" || loggedIn.SessionToken == "" {
		t.Fatalf("Login() = %+v, error %v", loggedIn, err)
	}
	if loggedIn.SessionToken == registered.SessionToken {
		t.Fatal("Login() reused an existing session token")
	}
}

func TestManagerRejectsInvalidCredentialsAndSessions(t *testing.T) {
	repository := NewMemoryRepository()
	manager, err := NewManager(repository, repository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if _, err := manager.Register(
		context.Background(), "player-one", "player@example.com", "strong-password",
	); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := manager.Login(
		context.Background(), "player@example.com", "wrong-password",
	); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := manager.AuthenticateSession(
		context.Background(), "not-a-session",
	); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("AuthenticateSession() error = %v, want ErrInvalidSession", err)
	}
}

func TestManagerRejectsShortPassword(t *testing.T) {
	repository := NewMemoryRepository()
	manager, err := NewManager(repository, repository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if _, err := manager.Register(
		context.Background(), "player-one", "player@example.com", "short",
	); err == nil {
		t.Fatal("Register() accepted a short password")
	}
}
