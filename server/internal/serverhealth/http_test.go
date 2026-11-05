package serverhealth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type checkerFunc func(context.Context) error

func (check checkerFunc) Ping(ctx context.Context) error {
	return check(ctx)
}

func TestLivenessIsHealthy(t *testing.T) {
	recorder := httptest.NewRecorder()
	Liveness().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("liveness response = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestReadinessWithoutDatabaseIsReady(t *testing.T) {
	recorder := httptest.NewRecorder()
	Readiness(nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "{\"status\":\"ready\"}\n" {
		t.Fatalf("readiness response = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestReadinessReportsUnavailableDatabase(t *testing.T) {
	recorder := httptest.NewRecorder()
	Readiness(checkerFunc(func(context.Context) error {
		return errors.New("database unavailable")
	})).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable ||
		recorder.Body.String() != "{\"status\":\"unavailable\"}\n" {
		t.Fatalf("readiness response = %d %q", recorder.Code, recorder.Body.String())
	}
}
