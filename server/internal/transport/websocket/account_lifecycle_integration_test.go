package websockettransport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
	httpaccount "github.com/mywyau/arcana/server/internal/transport/httpaccount"
)

type accountAuthenticationResponse struct {
	PlayerID     string `json:"player_id"`
	Email        string `json:"email"`
	SessionToken string `json:"session_token"`
}

func TestAccountLifecycleAcrossHTTPAndWebSocket(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("account.NewManager() error = %v", err)
	}
	guestManager, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("identity.NewGuestManager() error = %v", err)
	}
	characters := character.NewMemoryRepository()
	websocketHandler, err := NewHandlerWithRepositoriesAndOptions(
		world,
		interval,
		guestManager,
		characters,
		HandlerOptions{AccountManager: accountManager, RequireAccounts: true},
	)
	if err != nil {
		t.Fatalf("NewHandlerWithRepositoriesAndOptions() error = %v", err)
	}
	accountHandler, err := httpaccount.NewHandler(accountManager, guestManager, characters)
	if err != nil {
		t.Fatalf("httpaccount.NewHandler() error = %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/ws", websocketHandler)
	accountHandler.RegisterRoutes(mux)
	server := startTestServer(t, mux)
	defer server.Close()

	const email = "lifecycle@example.com"
	const password = "strong-password"
	registered := requestAccountAuthentication(
		t, server.URL+"/api/account/register", email, password,
	)
	if registered.PlayerID == "" || registered.SessionToken == "" {
		t.Fatalf("registration response = %+v", registered)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	connection := connectWithSession(
		t, ctx, websocketURL(server.URL)+"/ws", registered.SessionToken,
	)
	welcome := readServerMessage(t, ctx, connection)
	if welcome.Type != protocol.MessageWelcome || welcome.PlayerID != registered.PlayerID {
		t.Fatalf("registration session welcome = %+v", welcome)
	}
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type: protocol.MessageLogoutAccount, Sequence: 1,
	}); err != nil {
		t.Fatalf("write logout: %v", err)
	}
	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type == protocol.MessageAccountLoggedOut {
			break
		}
	}
	connection.CloseNow()
	waitForPlayerToLeave(t, world, simulation.PlayerID(registered.PlayerID))

	revokedConnection := connectWithSession(
		t, ctx, websocketURL(server.URL)+"/ws", registered.SessionToken,
	)
	revoked := readServerMessage(t, ctx, revokedConnection)
	revokedConnection.CloseNow()
	if revoked.Type != protocol.MessageError || revoked.Error == nil ||
		revoked.Error.Code != "invalid_account_session" {
		t.Fatalf("revoked session response = %+v", revoked)
	}

	loggedIn := requestAccountAuthentication(
		t, server.URL+"/api/account/login", email, password,
	)
	if loggedIn.PlayerID != registered.PlayerID || loggedIn.SessionToken == "" {
		t.Fatalf("login response = %+v, registration = %+v", loggedIn, registered)
	}
	if loggedIn.SessionToken == registered.SessionToken {
		t.Fatal("login reused the revoked session token")
	}

	reconnected := connectWithSession(
		t, ctx, websocketURL(server.URL)+"/ws", loggedIn.SessionToken,
	)
	defer reconnected.CloseNow()
	reconnectedWelcome := readServerMessage(t, ctx, reconnected)
	if reconnectedWelcome.Type != protocol.MessageWelcome ||
		reconnectedWelcome.PlayerID != registered.PlayerID {
		t.Fatalf("replacement session welcome = %+v", reconnectedWelcome)
	}
}

func requestAccountAuthentication(
	t *testing.T,
	url string,
	email string,
	password string,
) accountAuthenticationResponse {
	t.Helper()
	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, url, bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("POST %s error = %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("POST %s status = %s", url, response.Status)
	}
	var authentication accountAuthenticationResponse
	if err := json.NewDecoder(response.Body).Decode(&authentication); err != nil {
		t.Fatalf("decode POST %s response: %v", url, err)
	}
	return authentication
}

func connectWithSession(
	t *testing.T,
	ctx context.Context,
	url string,
	sessionToken string,
) *websocket.Conn {
	t.Helper()
	connection, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("websocket.Dial(%s) error = %v", url, err)
	}
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type: protocol.MessageAuthenticate, ProtocolVersion: protocol.CurrentVersion,
		SessionToken: sessionToken,
	}); err != nil {
		connection.CloseNow()
		t.Fatalf("authenticate WebSocket session: %v", err)
	}
	return connection
}

func (response accountAuthenticationResponse) String() string {
	return fmt.Sprintf(
		"{PlayerID:%q Email:%q SessionTokenPresent:%t}",
		response.PlayerID,
		response.Email,
		response.SessionToken != "",
	)
}
