package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type ctxKey int

const sessionCtxKey ctxKey = 1

// Middleware authenticates requests via the encrypted session cookie.
type Middleware struct {
	sess   *SessionStore
	logger *slog.Logger
}

// NewMiddleware constructs auth middleware.
func NewMiddleware(sess *SessionStore, logger *slog.Logger) *Middleware {
	return &Middleware{sess: sess, logger: logger}
}

// Require is a chi-compatible middleware that rejects unauthenticated
// requests with 401.
func (m *Middleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := m.sess.Load(r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		ctx := context.WithValue(r.Context(), sessionCtxKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth wraps a single http.HandlerFunc with auth.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	wrapped := m.Require(next)
	return wrapped.ServeHTTP
}

// SessionFromContext returns the authenticated session, or nil if absent.
func SessionFromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(sessionCtxKey).(*Session)
	return v
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
