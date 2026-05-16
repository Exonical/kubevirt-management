// Package config loads runtime configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AuthMode selects how user identity is propagated to the Kubernetes API.
type AuthMode string

const (
	// AuthModeOIDC forwards the user's OIDC ID token as the Bearer token to
	// kube-apiserver. The cluster must trust the same OIDC issuer.
	AuthModeOIDC AuthMode = "oidc"
	// AuthModeImpersonation uses the pod ServiceAccount token to call
	// kube-apiserver and sets Impersonate-User / Impersonate-Group headers
	// derived from the user's OIDC identity.
	AuthModeImpersonation AuthMode = "impersonation"
)

// Config holds all runtime configuration.
type Config struct {
	HTTPAddr string

	// Frontend / dev
	DevProxyTarget string // optional Vite dev server URL; if set, /assets and / are proxied here

	// Session cookie
	SessionHashKey  []byte
	SessionBlockKey []byte
	SessionName     string
	SessionMaxAge   time.Duration
	SessionSecure   bool

	// OIDC
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string
	OIDCUsernameClaim string
	OIDCGroupsClaim   string

	// Auth mode
	AuthMode             AuthMode
	ImpersonationEnabled bool

	// Kubernetes
	Kubeconfig         string // path; empty means in-cluster
	APIServerOverride  string // optional override of master URL
	InsecureSkipVerify bool
	CABundlePath       string
}

// Load reads configuration from the process environment.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:          envOr("HTTP_ADDR", ":8080"),
		DevProxyTarget:    os.Getenv("DEV_PROXY_TARGET"),
		SessionName:       envOr("SESSION_NAME", "kvm_session"),
		SessionMaxAge:     envDuration("SESSION_MAX_AGE", 12*time.Hour),
		SessionSecure:     envBool("SESSION_SECURE", true),
		OIDCIssuer:        os.Getenv("OIDC_ISSUER"),
		OIDCClientID:      os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:  os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:   os.Getenv("OIDC_REDIRECT_URL"),
		OIDCScopes:        envStringSlice("OIDC_SCOPES", []string{"openid", "profile", "email", "groups", "offline_access"}),
		OIDCUsernameClaim: envOr("OIDC_USERNAME_CLAIM", "email"),
		OIDCGroupsClaim:   envOr("OIDC_GROUPS_CLAIM", "groups"),
		AuthMode:          AuthMode(strings.ToLower(envOr("AUTH_MODE", string(AuthModeOIDC)))),
		ImpersonationEnabled: envBool("IMPERSONATION_ENABLED", false),
		Kubeconfig:         os.Getenv("KUBECONFIG"),
		APIServerOverride:  os.Getenv("KUBE_APISERVER"),
		InsecureSkipVerify: envBool("KUBE_INSECURE_SKIP_VERIFY", false),
		CABundlePath:       os.Getenv("KUBE_CA_BUNDLE"),
	}

	hashKey, err := envKey("SESSION_HASH_KEY", 64)
	if err != nil {
		return nil, fmt.Errorf("SESSION_HASH_KEY: %w", err)
	}
	cfg.SessionHashKey = hashKey

	blockKey, err := envKey("SESSION_BLOCK_KEY", 32)
	if err != nil {
		return nil, fmt.Errorf("SESSION_BLOCK_KEY: %w", err)
	}
	cfg.SessionBlockKey = blockKey

	switch cfg.AuthMode {
	case AuthModeOIDC, AuthModeImpersonation:
	default:
		return nil, fmt.Errorf("AUTH_MODE must be one of: oidc, impersonation (got %q)", cfg.AuthMode)
	}

	if cfg.AuthMode == AuthModeImpersonation {
		cfg.ImpersonationEnabled = true
	}

	if cfg.OIDCIssuer == "" || cfg.OIDCClientID == "" || cfg.OIDCRedirectURL == "" {
		return nil, errors.New("OIDC_ISSUER, OIDC_CLIENT_ID, and OIDC_REDIRECT_URL are required")
	}

	return cfg, nil
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func envStringSlice(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

// envKey reads a hex-encoded key from the environment. If absent and we're
// not running with SESSION_SECURE=true, it returns an error explaining how
// to generate one. In dev mode, callers may set DEV_SESSION_KEYS=true to
// auto-generate ephemeral keys (not done here; production must supply them).
func envKey(key string, sizeBytes int) ([]byte, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return nil, fmt.Errorf("required; generate with: openssl rand -hex %d", sizeBytes)
	}
	b, err := decodeHex(v)
	if err != nil {
		return nil, fmt.Errorf("must be hex-encoded: %w", err)
	}
	if len(b) != sizeBytes {
		return nil, fmt.Errorf("must decode to %d bytes (got %d)", sizeBytes, len(b))
	}
	return b, nil
}
