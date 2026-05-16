package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Exonical/kubevirt-management/internal/config"
)

func newTestConfig() *config.Config {
	return &config.Config{
		SessionHashKey:  []byte(strings.Repeat("a", 64)),
		SessionBlockKey: []byte(strings.Repeat("b", 32)),
		SessionName:     "test_session",
		SessionMaxAge:   time.Hour,
		SessionSecure:   false,
	}
}

func TestSessionRoundTrip(t *testing.T) {
	store := NewSessionStore(newTestConfig())
	w := httptest.NewRecorder()
	in := &Session{
		Subject:   "abc",
		Username:  "alice@example.com",
		Groups:    []string{"devs"},
		IDToken:   "token",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}
	if err := store.Save(w, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	out, err := store.Load(req)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.Subject != in.Subject || out.Username != in.Username || out.IDToken != in.IDToken {
		t.Errorf("round trip mismatch: %+v", out)
	}
	if len(out.Groups) != 1 || out.Groups[0] != "devs" {
		t.Errorf("groups mismatch: %v", out.Groups)
	}
}

func TestSessionExpired(t *testing.T) {
	s := &Session{ExpiresAt: time.Now().Add(-time.Minute)}
	if !s.Expired() {
		t.Error("expected expired session")
	}
	s.ExpiresAt = time.Now().Add(time.Minute)
	if s.Expired() {
		t.Error("expected non-expired session")
	}
}
