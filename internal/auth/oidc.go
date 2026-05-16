package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/Exonical/kubevirt-management/internal/config"
)

// OIDCManager handles discovery, login redirects, and token exchange.
type OIDCManager struct {
	cfg      *config.Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
	logger   *slog.Logger
}

// NewOIDCManager performs OIDC discovery against the configured issuer.
func NewOIDCManager(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*OIDCManager, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})

	return &OIDCManager{
		cfg:      cfg,
		provider: provider,
		verifier: verifier,
		oauth: &oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.OIDCRedirectURL,
			Scopes:       cfg.OIDCScopes,
		},
		logger: logger,
	}, nil
}

// AuthCodeURL returns the URL the client should be redirected to in order
// to begin the OIDC flow. State and nonce are random opaque values that
// the callback handler will verify.
func (m *OIDCManager) AuthCodeURL(state, nonce string) string {
	return m.oauth.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.AccessTypeOffline)
}

// Exchange swaps an authorization code for tokens and verifies the ID
// token, returning a populated Session.
func (m *OIDCManager) Exchange(ctx context.Context, code, nonce string) (*Session, error) {
	tok, err := m.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange: %w", err)
	}
	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, errors.New("id_token missing from token response")
	}
	idTok, err := m.verifier.Verify(ctx, rawID)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}
	if idTok.Nonce != nonce {
		return nil, errors.New("nonce mismatch")
	}

	var claims map[string]any
	if err := idTok.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}

	username, _ := claims[m.cfg.OIDCUsernameClaim].(string)
	if username == "" {
		if email, ok := claims["email"].(string); ok {
			username = email
		} else {
			username = idTok.Subject
		}
	}
	email, _ := claims["email"].(string)
	groups := extractGroups(claims, m.cfg.OIDCGroupsClaim)

	expiry := idTok.Expiry
	if expiry.IsZero() {
		expiry = time.Now().Add(time.Hour)
	}

	return &Session{
		Subject:      idTok.Subject,
		Username:     username,
		Email:        email,
		Groups:       groups,
		IDToken:      rawID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiry,
		IssuedAt:     time.Now(),
	}, nil
}

func extractGroups(claims map[string]any, claim string) []string {
	raw, ok := claims[claim]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, g := range v {
			if s, ok := g.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	case string:
		return []string{v}
	default:
		return nil
	}
}

// Handler exposes HTTP handlers for the OIDC login flow.
type Handler struct {
	cfg    *config.Config
	mgr    *OIDCManager
	sess   *SessionStore
	state  *stateStore
	logger *slog.Logger
}

// NewHandler constructs the auth Handler.
func NewHandler(cfg *config.Config, mgr *OIDCManager, sess *SessionStore, logger *slog.Logger) *Handler {
	return &Handler{
		cfg:    cfg,
		mgr:    mgr,
		sess:   sess,
		state:  newStateStore(cfg),
		logger: logger,
	}
}

// Login starts the OIDC flow by redirecting to the provider.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := randomString(32)
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	nonce, err := randomString(32)
	if err != nil {
		http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
		return
	}
	redirectTo := r.URL.Query().Get("redirect_to")
	if redirectTo == "" {
		redirectTo = "/"
	}
	if err := h.state.save(w, &stateData{State: state, Nonce: nonce, RedirectTo: redirectTo}); err != nil {
		http.Error(w, "failed to save state", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.mgr.AuthCodeURL(state, nonce), http.StatusFound)
}

// Callback completes the OIDC flow.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	sd, err := h.state.load(r)
	if err != nil {
		http.Error(w, "missing or invalid state cookie", http.StatusBadRequest)
		return
	}
	h.state.clear(w)
	if r.URL.Query().Get("state") != sd.State {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, "oidc error: "+errParam, http.StatusUnauthorized)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	sess, err := h.mgr.Exchange(r.Context(), code, sd.Nonce)
	if err != nil {
		h.logger.Error("oidc exchange failed", "err", err)
		http.Error(w, "oidc exchange failed", http.StatusUnauthorized)
		return
	}
	if err := h.sess.Save(w, sess); err != nil {
		h.logger.Error("save session", "err", err)
		http.Error(w, "save session failed", http.StatusInternalServerError)
		return
	}
	target := sd.RedirectTo
	if target == "" {
		target = "/"
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// Logout clears the session cookie. The OIDC provider is not contacted;
// front-channel logout can be added later if needed.
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	h.sess.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me returns the authenticated user's identity.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	sess := SessionFromContext(r.Context())
	if sess == nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	payload := map[string]any{
		"subject":  sess.Subject,
		"username": sess.Username,
		"email":    sess.Email,
		"groups":   sess.Groups,
		"expires_at": sess.ExpiresAt,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
