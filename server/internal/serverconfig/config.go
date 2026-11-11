package serverconfig

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	EnvironmentDevelopment  = "development"
	EnvironmentStaging      = "staging"
	EnvironmentProduction   = "production"
	DefaultVisibilityRadius = 40.0
)

type Config struct {
	Environment                string
	ListenAddress              string
	DatabaseURL                string
	DebugContentEnabled        bool
	TestSupplyChestsEnabled    bool
	GuestAuthenticationEnabled bool
	TrustProxyHeaders          bool
	AllowedOrigins             []string
	VisibilityRadius           float64
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
	testSupplyChestsEnabled := environment == EnvironmentDevelopment
	if rawChests := strings.TrimSpace(getenv("ARCANA_ENABLE_TEST_SUPPLY_CHESTS")); rawChests != "" {
		testSupplyChestsEnabled, err = strconv.ParseBool(rawChests)
		if err != nil {
			return Config{}, fmt.Errorf("ARCANA_ENABLE_TEST_SUPPLY_CHESTS must be true or false")
		}
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
	trustProxyHeaders := false
	if rawTrustProxy := strings.TrimSpace(getenv("ARCANA_TRUST_PROXY_HEADERS")); rawTrustProxy != "" {
		trustProxyHeaders, err = strconv.ParseBool(rawTrustProxy)
		if err != nil {
			return Config{}, fmt.Errorf("ARCANA_TRUST_PROXY_HEADERS must be true or false")
		}
	}
	visibilityRadius := DefaultVisibilityRadius
	if rawRadius := strings.TrimSpace(getenv("ARCANA_VISIBILITY_RADIUS")); rawRadius != "" {
		visibilityRadius, err = strconv.ParseFloat(rawRadius, 64)
		if err != nil || math.IsNaN(visibilityRadius) || math.IsInf(visibilityRadius, 0) || visibilityRadius <= 0 {
			return Config{}, fmt.Errorf("ARCANA_VISIBILITY_RADIUS must be a finite positive number")
		}
	}
	return Config{
		Environment:                environment,
		ListenAddress:              ":" + port,
		DatabaseURL:                databaseURL,
		DebugContentEnabled:        debugContentEnabled,
		TestSupplyChestsEnabled:    testSupplyChestsEnabled,
		GuestAuthenticationEnabled: guestAuthenticationEnabled,
		TrustProxyHeaders:          trustProxyHeaders,
		AllowedOrigins:             splitList(getenv("ARCANA_ALLOWED_ORIGINS")),
		VisibilityRadius:           visibilityRadius,
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
