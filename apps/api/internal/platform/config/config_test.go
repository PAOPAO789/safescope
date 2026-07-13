package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("API_PORT", "")
	t.Setenv("JWT_TTL", "")
	t.Setenv("APP_ENV", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.JWTTTL.Hours() != 24 {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestProductionRequiresStrongSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error")
	}
}
