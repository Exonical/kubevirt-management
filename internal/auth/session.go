// Package auth handles OIDC login, encrypted session cookies, and
// propagation of user identity to the Kubernetes API.
package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"

	"github.com/Exonical/kubevirt-management/internal/config"
)

// Session is the data we persist in the encrypted session cookie.
//
// We deliberately store the OIDC tokens directly so we can forward them to
// kube-apiserver on every request without server-side state. ID tokens are
// typically a few KB; if this grows beyond the 4KB cookie limit we'll need
// to introduce server-side session storage.
type Session struct {
	Subject      string    `json:"sub"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	Groups       []string  `json:"groups,omitempty"`
	IDToken      string    `json:"id_token"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	IssuedAt     time.Time `json:"issued_at"`
}

// Expired reports whether the session's ID token is past its expiry.
func (s *Session) Expired() bool {
	return !s.ExpiresAt.IsZero() && time.Now().After(s.ExpiresAt)
}

// SessionStore encodes and decodes Session values to/from cookies.
type SessionStore struct {
	cookieName string
	maxAge     time.Duration
	secure     bool
	sc         *securecookie.SecureCookie
}

// NewSessionStore constructs a SessionStore from config.
func NewSessionStore(cfg *config.Config) *SessionStore {
	sc := securecookie.New(cfg.SessionHashKey, cfg.SessionBlockKey)
	sc.SetSerializer(securecookie.JSONEncoder{})
	sc.MaxAge(int(cfg.SessionMaxAge.Seconds()))
	return &SessionStore{
		cookieName: cfg.SessionName,
		maxAge:     cfg.SessionMaxAge,
		secure:     cfg.SessionSecure,
		sc:         sc,
	}
}

// Save serializes the session into an encrypted cookie on the response.
func (s *SessionStore) Save(w http.ResponseWriter, sess *Session) error {
	if sess == nil {
		return errors.New("session is nil")
	}
	encoded, err := s.sc.Encode(s.cookieName, sess)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.maxAge.Seconds()),
	})
	return nil
}

// Load reads the session from the request, or returns an error if absent /
// invalid / expired.
func (s *SessionStore) Load(r *http.Request) (*Session, error) {
	c, err := r.Cookie(s.cookieName)
	if err != nil {
		return nil, err
	}
	var sess Session
	if err := s.sc.Decode(s.cookieName, c.Value, &sess); err != nil {
		return nil, err
	}
	if sess.Expired() {
		return nil, errors.New("session expired")
	}
	return &sess, nil
}

// Clear writes an expired cookie to the response, effectively logging the
// user out.
func (s *SessionStore) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// stateStore manages the short-lived OIDC state/nonce cookie used during
// the login flow. We keep it separate from the main session so an in-flight
// login does not overwrite an existing session.
type stateStore struct {
	name   string
	secure bool
	sc     *securecookie.SecureCookie
}

func newStateStore(cfg *config.Config) *stateStore {
	sc := securecookie.New(cfg.SessionHashKey, cfg.SessionBlockKey)
	sc.SetSerializer(securecookie.JSONEncoder{})
	sc.MaxAge(int((10 * time.Minute).Seconds()))
	return &stateStore{
		name:   cfg.SessionName + "_state",
		secure: cfg.SessionSecure,
		sc:     sc,
	}
}

type stateData struct {
	State       string `json:"state"`
	Nonce       string `json:"nonce"`
	RedirectTo  string `json:"redirect_to,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
}

func (s *stateStore) save(w http.ResponseWriter, sd *stateData) error {
	enc, err := s.sc.Encode(s.name, sd)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     s.name,
		Value:    enc,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((10 * time.Minute).Seconds()),
	})
	return nil
}

func (s *stateStore) load(r *http.Request) (*stateData, error) {
	c, err := r.Cookie(s.name)
	if err != nil {
		return nil, err
	}
	var sd stateData
	if err := s.sc.Decode(s.name, c.Value, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

func (s *stateStore) clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.name,
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
