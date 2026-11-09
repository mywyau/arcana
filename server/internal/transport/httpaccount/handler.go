package httpaccount

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/identity"
)

const maximumRequestBytes = 4 * 1024

type Handler struct {
	accounts   *account.Manager
	guests     *identity.GuestManager
	characters account.CharacterRepository
}

type credentialsRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	GuestToken string `json:"guest_token,omitempty"`
}

type authenticationResponse struct {
	PlayerID     string `json:"player_id"`
	Email        string `json:"email"`
	SessionToken string `json:"session_token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(
	accounts *account.Manager,
	guests *identity.GuestManager,
	characters account.CharacterRepository,
) (*Handler, error) {
	if accounts == nil {
		return nil, fmt.Errorf("account manager is required")
	}
	if guests == nil {
		return nil, fmt.Errorf("guest identity manager is required")
	}
	if characters == nil {
		return nil, fmt.Errorf("character repository is required")
	}
	return &Handler{accounts: accounts, guests: guests, characters: characters}, nil
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/account/register", handler.register)
	mux.HandleFunc("POST /api/account/login", handler.login)
	mux.HandleFunc("POST /api/account/logout", handler.logout)
}

func (handler *Handler) register(response http.ResponseWriter, request *http.Request) {
	credentials, ok := decodeCredentials(response, request)
	if !ok {
		return
	}
	var authentication account.Authentication
	var err error
	if credentials.GuestToken == "" {
		authentication, err = handler.accounts.RegisterNewCharacter(
			request.Context(), credentials.Email, credentials.Password, handler.characters,
		)
	} else {
		guest, guestErr := handler.guests.Authenticate(request.Context(), credentials.GuestToken)
		if guestErr != nil {
			writeError(response, http.StatusUnauthorized, "guest character could not be verified")
			return
		}
		authentication, err = handler.accounts.Register(
			request.Context(), guest.PlayerID, credentials.Email, credentials.Password,
		)
	}
	if err != nil {
		switch {
		case errors.Is(err, account.ErrAccountExists):
			writeError(response, http.StatusConflict, "that email address is already registered")
		case errors.Is(err, account.ErrCharacterUnavailable):
			writeError(response, http.StatusConflict, "that character is already linked to an account")
		case errors.Is(err, account.ErrInvalidEmail), errors.Is(err, account.ErrInvalidPassword):
			writeError(response, http.StatusBadRequest, err.Error())
		default:
			writeError(response, http.StatusInternalServerError, "could not create account")
		}
		return
	}
	writeJSON(response, http.StatusCreated, authenticationResponse{
		PlayerID: authentication.PlayerID, Email: authentication.Email,
		SessionToken: authentication.SessionToken,
	})
}

func (handler *Handler) login(response http.ResponseWriter, request *http.Request) {
	credentials, ok := decodeCredentials(response, request)
	if !ok {
		return
	}
	authentication, err := handler.accounts.Login(
		request.Context(), credentials.Email, credentials.Password,
	)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCredentials) {
			writeError(response, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeError(response, http.StatusInternalServerError, "could not log in")
		return
	}
	writeJSON(response, http.StatusOK, authenticationResponse{
		PlayerID: authentication.PlayerID, Email: authentication.Email,
		SessionToken: authentication.SessionToken,
	})
}

func (handler *Handler) logout(response http.ResponseWriter, request *http.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
	if token == "" || token == request.Header.Get("Authorization") {
		writeError(response, http.StatusUnauthorized, "account session is required")
		return
	}
	if err := handler.accounts.Logout(request.Context(), token); err != nil {
		if errors.Is(err, account.ErrInvalidSession) {
			writeError(response, http.StatusUnauthorized, "invalid or expired account session")
			return
		}
		writeError(response, http.StatusInternalServerError, "could not log out")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func decodeCredentials(
	response http.ResponseWriter,
	request *http.Request,
) (credentialsRequest, bool) {
	request.Body = http.MaxBytesReader(response, request.Body, maximumRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var credentials credentialsRequest
	if err := decoder.Decode(&credentials); err != nil {
		writeError(response, http.StatusBadRequest, "request body is invalid")
		return credentialsRequest{}, false
	}
	if strings.TrimSpace(credentials.Email) == "" || credentials.Password == "" {
		writeError(response, http.StatusBadRequest, "email and password are required")
		return credentialsRequest{}, false
	}
	return credentials, true
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, errorResponse{Error: message})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
