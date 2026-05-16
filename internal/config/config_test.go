package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresOIDC(t *testing.T) {
	t.Setenv("SESSION_HASH_KEY", strings.Repeat("a", 128))
	t.Setenv("SESSION_BLOCK_KEY", strings.Repeat("b", 64))
	if _, err := Load(); err == nil {
		t.Fatal("expected error when OIDC env is missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SESSION_HASH_KEY", strings.Repeat("a", 128))
	t.Setenv("SESSION_BLOCK_KEY", strings.Repeat("b", 64))
	t.Setenv("OIDC_ISSUER", "https://example.com")
	t.Setenv("OIDC_CLIENT_ID", "client")
	t.Setenv("OIDC_REDIRECT_URL", "http://localhost/cb")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.AuthMode != AuthModeOIDC {
		t.Errorf("AuthMode = %q, want oidc", cfg.AuthMode)
	}
	if cfg.OIDCUsernameClaim != "email" {
		t.Errorf("OIDCUsernameClaim = %q, want email", cfg.OIDCUsernameClaim)
	}
}

func TestLoadImpersonationMode(t *testing.T) {
	t.Setenv("SESSION_HASH_KEY", strings.Repeat("a", 128))
	t.Setenv("SESSION_BLOCK_KEY", strings.Repeat("b", 64))
	t.Setenv("OIDC_ISSUER", "https://example.com")
	t.Setenv("OIDC_CLIENT_ID", "client")
	t.Setenv("OIDC_REDIRECT_URL", "http://localhost/cb")
	t.Setenv("AUTH_MODE", "impersonation")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.ImpersonationEnabled {
		t.Error("ImpersonationEnabled = false, want true when AUTH_MODE=impersonation")
	}
}

func TestEnvStringSlice(t *testing.T) {
	def := []string{"a", "b"}
	t.Setenv("X", "")
	if got := envStringSlice("X", def); len(got) != 2 || got[0] != "a" {
		t.Errorf("empty env should return default, got %v", got)
	}
	t.Setenv("X", "  foo , bar ,, baz")
	got := envStringSlice("X", def)
	want := []string{"foo", "bar", "baz"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
