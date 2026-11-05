package serverhealth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const readinessTimeout = time.Second

type Checker interface {
	Ping(context.Context) error
}

func Liveness() http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		writeStatus(response, http.StatusOK, "ok")
	})
}

func Readiness(checker Checker) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if checker != nil {
			ctx, cancel := context.WithTimeout(request.Context(), readinessTimeout)
			defer cancel()
			if err := checker.Ping(ctx); err != nil {
				writeStatus(response, http.StatusServiceUnavailable, "unavailable")
				return
			}
		}
		writeStatus(response, http.StatusOK, "ready")
	})
}

func writeStatus(response http.ResponseWriter, statusCode int, status string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	_ = json.NewEncoder(response).Encode(map[string]string{"status": status})
}
