package auth

import (
	"testing"

	"k8s.io/client-go/rest"

	"github.com/Exonical/kubevirt-management/internal/config"
)

func TestApplyAuthOIDCMode(t *testing.T) {
	base := &rest.Config{BearerToken: "sa-token"}
	sess := &Session{Username: "alice", Groups: []string{"devs"}, IDToken: "id-token"}
	out := ApplyAuth(base, sess, config.AuthModeOIDC)
	if out.BearerToken != "id-token" {
		t.Errorf("BearerToken = %q, want id-token", out.BearerToken)
	}
	if out.Impersonate.UserName != "" {
		t.Errorf("Impersonate should be empty, got %q", out.Impersonate.UserName)
	}
}

func TestApplyAuthImpersonationMode(t *testing.T) {
	base := &rest.Config{BearerToken: "sa-token"}
	sess := &Session{Username: "alice", Groups: []string{"devs"}, IDToken: "id-token"}
	out := ApplyAuth(base, sess, config.AuthModeImpersonation)
	if out.BearerToken != "sa-token" {
		t.Errorf("BearerToken should remain sa-token, got %q", out.BearerToken)
	}
	if out.Impersonate.UserName != "alice" {
		t.Errorf("Impersonate.UserName = %q, want alice", out.Impersonate.UserName)
	}
	if len(out.Impersonate.Groups) != 1 || out.Impersonate.Groups[0] != "devs" {
		t.Errorf("Impersonate.Groups = %v", out.Impersonate.Groups)
	}
}
