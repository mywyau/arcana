package httpaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestAccountHTTPRegistrationLoginAndLogout(t *testing.T) {
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	guestManager, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() error = %v", err)
	}
	handler, err := NewHandler(accountManager, guestManager, character.NewMemoryRepository())
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	registered := requestAuthentication(t, mux, "/api/account/register", map[string]string{
		"email": "player@example.com", "password": "strong-password",
	})
	if registered.PlayerID == "" || registered.SessionToken == "" {
		t.Fatalf("registration response = %+v", registered)
	}
	loggedIn := requestAuthentication(t, mux, "/api/account/login", map[string]string{
		"email": "PLAYER@example.com", "password": "strong-password",
	})
	if loggedIn.PlayerID != registered.PlayerID || loggedIn.SessionToken == "" {
		t.Fatalf("login response = %+v, registration = %+v", loggedIn, registered)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/account/logout", nil)
	request.Header.Set("Authorization", "Bearer "+loggedIn.SessionToken)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", response.Code)
	}
	if _, err := accountManager.AuthenticateSession(
		context.Background(), loggedIn.SessionToken,
	); err == nil {
		t.Fatal("logged-out session still authenticates")
	}
	if _, err := accountManager.AuthenticateSession(
		context.Background(), registered.SessionToken,
	); err != nil {
		t.Fatalf("registration session was also revoked: %v", err)
	}
}

func TestAccountHTTPRegistrationLinksExistingGuest(t *testing.T) {
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	guestManager, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() error = %v", err)
	}
	guest, err := guestManager.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("create guest: %v", err)
	}
	characters := character.NewMemoryRepository()
	if err := characters.Save(
		context.Background(), guest.PlayerID, simulation.DefaultPersistentPlayerState(),
	); err != nil {
		t.Fatalf("save guest character: %v", err)
	}
	handler, err := NewHandler(accountManager, guestManager, characters)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	registered := requestAuthentication(t, mux, "/api/account/register", map[string]string{
		"email": "guest@example.com", "password": "strong-password", "guest_token": guest.Token,
	})
	if registered.PlayerID != guest.PlayerID {
		t.Fatalf("registered player = %s, want %s", registered.PlayerID, guest.PlayerID)
	}
}

func requestAuthentication(
	t *testing.T,
	handler http.Handler,
	path string,
	payload map[string]string,
) authenticationResponse {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code < 200 || response.Code >= 300 {
		t.Fatalf("POST %s status = %d", path, response.Code)
	}
	var authentication authenticationResponse
	if err := json.NewDecoder(response.Body).Decode(&authentication); err != nil {
		t.Fatalf("decode authentication response: %v", err)
	}
	return authentication
}
