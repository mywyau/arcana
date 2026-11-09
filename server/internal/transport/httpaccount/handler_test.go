package httpaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/ratelimit"
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

func TestLoginRateLimitBlocksFailuresAndRecovers(t *testing.T) {
	now := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	characters := character.NewMemoryRepository()
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if _, err := accountManager.RegisterNewCharacter(
		context.Background(), "player@example.com", "strong-password", characters,
	); err != nil {
		t.Fatalf("RegisterNewCharacter() error = %v", err)
	}
	handler := newRateLimitedHandler(t, accountManager, characters, HandlerOptions{
		LoginAttemptLimiter: testLimiter(t, 20, time.Minute, &now),
		LoginFailureLimiter: testLimiter(t, 2, time.Minute, &now),
	})

	for attempt := 0; attempt < 2; attempt++ {
		response := requestJSON(t, handler, "/api/account/login", map[string]string{
			"email": "player@example.com", "password": "wrong-password",
		}, "192.0.2.10:1234", "")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("failed login %d status = %d", attempt+1, response.Code)
		}
	}
	limited := requestJSON(t, handler, "/api/account/login", map[string]string{
		"email": "player@example.com", "password": "strong-password",
	}, "192.0.2.10:1234", "")
	if limited.Code != http.StatusTooManyRequests || limited.Header().Get("Retry-After") != "60" {
		t.Fatalf("limited login status = %d, Retry-After = %q", limited.Code, limited.Header().Get("Retry-After"))
	}

	now = now.Add(time.Minute)
	recovered := requestJSON(t, handler, "/api/account/login", map[string]string{
		"email": "player@example.com", "password": "strong-password",
	}, "192.0.2.10:1234", "")
	if recovered.Code != http.StatusOK {
		t.Fatalf("login after cooldown status = %d", recovered.Code)
	}
}

func TestSuccessfulLoginClearsOnlyThatEmailFailureHistory(t *testing.T) {
	now := time.Now()
	characters := character.NewMemoryRepository()
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if _, err := accountManager.RegisterNewCharacter(
		context.Background(), "player@example.com", "strong-password", characters,
	); err != nil {
		t.Fatalf("RegisterNewCharacter() error = %v", err)
	}
	handler := newRateLimitedHandler(t, accountManager, characters, HandlerOptions{
		LoginAttemptLimiter: testLimiter(t, 20, time.Minute, &now),
		LoginFailureLimiter: testLimiter(t, 2, time.Minute, &now),
	})

	wrong := map[string]string{"email": "player@example.com", "password": "wrong-password"}
	if response := requestJSON(t, handler, "/api/account/login", wrong, "192.0.2.20:1234", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("first failed login status = %d", response.Code)
	}
	correct := map[string]string{"email": "player@example.com", "password": "strong-password"}
	if response := requestJSON(t, handler, "/api/account/login", correct, "192.0.2.20:1234", ""); response.Code != http.StatusOK {
		t.Fatalf("successful login status = %d", response.Code)
	}
	if response := requestJSON(t, handler, "/api/account/login", wrong, "192.0.2.20:1234", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("failure after success status = %d", response.Code)
	}
}

func TestRegistrationRateLimitUsesTrustedForwardedAddress(t *testing.T) {
	now := time.Now()
	characters := character.NewMemoryRepository()
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	handler := newRateLimitedHandler(t, accountManager, characters, HandlerOptions{
		RegistrationLimiter: testLimiter(t, 1, time.Minute, &now),
		TrustProxyHeaders:   true,
	})

	for index, forwardedAddress := range []string{"198.51.100.1, 10.0.0.1", "198.51.100.2, 10.0.0.1"} {
		response := requestJSON(t, handler, "/api/account/register", map[string]string{
			"email": "player" + strconv.Itoa(index) + "@example.com", "password": "strong-password",
		}, "10.0.0.1:1234", forwardedAddress)
		if response.Code != http.StatusCreated {
			t.Fatalf("forwarded client %d registration status = %d", index+1, response.Code)
		}
	}
	limited := requestJSON(t, handler, "/api/account/register", map[string]string{
		"email": "third@example.com", "password": "strong-password",
	}, "10.0.0.1:1234", "198.51.100.1, 10.0.0.1")
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("repeated forwarded client registration status = %d", limited.Code)
	}
}

func TestRegistrationRateLimitIgnoresUntrustedForwardedAddress(t *testing.T) {
	now := time.Now()
	characters := character.NewMemoryRepository()
	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	handler := newRateLimitedHandler(t, accountManager, characters, HandlerOptions{
		RegistrationLimiter: testLimiter(t, 1, time.Minute, &now),
	})

	first := requestJSON(t, handler, "/api/account/register", map[string]string{
		"email": "first@example.com", "password": "strong-password",
	}, "10.0.0.1:1234", "198.51.100.1")
	if first.Code != http.StatusCreated {
		t.Fatalf("first registration status = %d", first.Code)
	}
	limited := requestJSON(t, handler, "/api/account/register", map[string]string{
		"email": "second@example.com", "password": "strong-password",
	}, "10.0.0.1:1234", "198.51.100.2")
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("spoofed forwarded client registration status = %d", limited.Code)
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

func newRateLimitedHandler(
	t *testing.T,
	accountManager *account.Manager,
	characters account.CharacterRepository,
	options HandlerOptions,
) http.Handler {
	t.Helper()
	guestManager, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() error = %v", err)
	}
	handler, err := NewHandlerWithOptions(accountManager, guestManager, characters, options)
	if err != nil {
		t.Fatalf("NewHandlerWithOptions() error = %v", err)
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}

func testLimiter(
	t *testing.T,
	limit int,
	window time.Duration,
	now *time.Time,
) *ratelimit.Limiter {
	t.Helper()
	limiter, err := ratelimit.New(ratelimit.Config{
		Limit: limit, Window: window, MaxEntries: 100,
		Now: func() time.Time { return *now },
	})
	if err != nil {
		t.Fatalf("ratelimit.New() error = %v", err)
	}
	return limiter
}

func requestJSON(
	t *testing.T,
	handler http.Handler,
	path string,
	payload map[string]string,
	remoteAddress string,
	forwardedAddress string,
) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.RemoteAddr = remoteAddress
	request.Header.Set("Content-Type", "application/json")
	if forwardedAddress != "" {
		request.Header.Set("X-Forwarded-For", forwardedAddress)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
