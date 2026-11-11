package serverconfig

import "testing"

func TestLoadDefaultsToLocalDevelopment(t *testing.T) {
	config, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Environment != EnvironmentDevelopment || config.ListenAddress != ":8080" ||
		!config.DebugContentEnabled || !config.TestSupplyChestsEnabled || !config.GuestAuthenticationEnabled ||
		config.VisibilityRadius != DefaultVisibilityRadius {
		t.Fatalf("development config = %+v", config)
	}
}

func TestLoadProductionConfiguration(t *testing.T) {
	values := map[string]string{
		"ARCANA_ENV":                       "production",
		"PORT":                             "10000",
		"DATABASE_URL":                     "postgres://internal/arcana",
		"ARCANA_ALLOWED_ORIGINS":           "https://game.example.com, launcher.example.com",
		"ARCANA_TRUST_PROXY_HEADERS":       "true",
		"ARCANA_ENABLE_TEST_SUPPLY_CHESTS": "true",
		"ARCANA_VISIBILITY_RADIUS":         "55.5",
	}
	config, err := Load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.ListenAddress != ":10000" || config.DebugContentEnabled || !config.TestSupplyChestsEnabled ||
		config.GuestAuthenticationEnabled || !config.TrustProxyHeaders || len(config.AllowedOrigins) != 2 ||
		config.VisibilityRadius != 55.5 {
		t.Fatalf("production config = %+v", config)
	}
}

func TestLoadRejectsUnsafeProductionConfiguration(t *testing.T) {
	tests := []map[string]string{
		{"ARCANA_ENV": "production"},
		{"ARCANA_ENV": "production", "DATABASE_URL": "postgres://db", "ARCANA_ENABLE_DEBUG_CONTENT": "true"},
		{"ARCANA_ENV": "production", "DATABASE_URL": "postgres://db", "ARCANA_ALLOW_GUEST_AUTH": "true"},
		{"ARCANA_ALLOW_GUEST_AUTH": "invalid"},
		{"ARCANA_ENABLE_TEST_SUPPLY_CHESTS": "invalid"},
		{"ARCANA_TRUST_PROXY_HEADERS": "invalid"},
		{"ARCANA_VISIBILITY_RADIUS": "0"},
		{"ARCANA_VISIBILITY_RADIUS": "NaN"},
		{"PORT": "invalid"},
	}
	for _, values := range tests {
		if _, err := Load(func(key string) string { return values[key] }); err == nil {
			t.Fatalf("Load(%v) error = nil", values)
		}
	}
}
