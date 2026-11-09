package serverconfig

import "testing"

func TestLoadDefaultsToLocalDevelopment(t *testing.T) {
	config, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Environment != EnvironmentDevelopment || config.ListenAddress != ":8080" ||
		!config.DebugContentEnabled || !config.GuestAuthenticationEnabled {
		t.Fatalf("development config = %+v", config)
	}
}

func TestLoadProductionConfiguration(t *testing.T) {
	values := map[string]string{
		"ARCANA_ENV":             "production",
		"PORT":                   "10000",
		"DATABASE_URL":           "postgres://internal/arcana",
		"ARCANA_ALLOWED_ORIGINS": "https://game.example.com, launcher.example.com",
	}
	config, err := Load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.ListenAddress != ":10000" || config.DebugContentEnabled ||
		config.GuestAuthenticationEnabled || len(config.AllowedOrigins) != 2 {
		t.Fatalf("production config = %+v", config)
	}
}

func TestLoadRejectsUnsafeProductionConfiguration(t *testing.T) {
	tests := []map[string]string{
		{"ARCANA_ENV": "production"},
		{"ARCANA_ENV": "production", "DATABASE_URL": "postgres://db", "ARCANA_ENABLE_DEBUG_CONTENT": "true"},
		{"ARCANA_ENV": "production", "DATABASE_URL": "postgres://db", "ARCANA_ALLOW_GUEST_AUTH": "true"},
		{"ARCANA_ALLOW_GUEST_AUTH": "invalid"},
		{"PORT": "invalid"},
	}
	for _, values := range tests {
		if _, err := Load(func(key string) string { return values[key] }); err == nil {
			t.Fatalf("Load(%v) error = nil", values)
		}
	}
}
