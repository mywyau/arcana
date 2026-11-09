package serverconfig

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"
)

type Config struct {
	Environment                string
	ListenAddress              string
	DatabaseURL                string
	DebugContentEnabled        bool
	GuestAuthenticationEnabled bool
	AllowedOrigins             []string
}

func Load(getenv func(string) string) (Config, error) {
	environment := strings.ToLower(strings.TrimSpace(getenv("ARCANA_ENV")))
	if environment == "" {
		environment = EnvironmentDevelopment
	}
	if environment != EnvironmentDevelopment && environment != EnvironmentStaging && environment != EnvironmentProduction {
		return Config{}, fmt.Errorf("ARCANA_ENV must be development, staging, or production")
	}
	port := strings.TrimSpace(getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}
	databaseURL := strings.TrimSpace(getenv("DATABASE_URL"))
	if environment != EnvironmentDevelopment && databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required outside development")
	}
	debugContentEnabled := environment == EnvironmentDevelopment
	if rawDebug := strings.TrimSpace(getenv("ARCANA_ENABLE_DEBUG_CONTENT")); rawDebug != "" {
		debugContentEnabled, err = strconv.ParseBool(rawDebug)
		if err != nil {
			return Config{}, fmt.Errorf("ARCANA_ENABLE_DEBUG_CONTENT must be true or false")
		}
	}
	if environment == EnvironmentProduction && debugContentEnabled {
		return Config{}, fmt.Errorf("debug content cannot be enabled in production")
	}
	guestAuthenticationEnabled := environment == EnvironmentDevelopment
	if rawGuests := strings.TrimSpace(getenv("ARCANA_ALLOW_GUEST_AUTH")); rawGuests != "" {
		guestAuthenticationEnabled, err = strconv.ParseBool(rawGuests)
		if err != nil {
			return Config{}, fmt.Errorf("ARCANA_ALLOW_GUEST_AUTH must be true or false")
		}
	}
	if environment == EnvironmentProduction && guestAuthenticationEnabled {
		return Config{}, fmt.Errorf("guest authentication cannot be enabled in production")
	}
	return Config{
		Environment:                environment,
		ListenAddress:              ":" + port,
		DatabaseURL:                databaseURL,
		DebugContentEnabled:        debugContentEnabled,
		GuestAuthenticationEnabled: guestAuthenticationEnabled,
		AllowedOrigins:             splitList(getenv("ARCANA_ALLOWED_ORIGINS")),
	}, nil
}

func splitList(value string) []string {
	var values []string
	for _, entry := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(entry); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
