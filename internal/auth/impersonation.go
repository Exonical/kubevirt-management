package auth

import (
	"k8s.io/client-go/rest"

	"github.com/Exonical/kubevirt-management/internal/config"
)

// ApplyAuth configures a *rest.Config to act as the authenticated user.
//
// In OIDC mode, the user's ID token is set as the Bearer token; the
// kube-apiserver must trust the same OIDC issuer to validate it.
//
// In impersonation mode, the base config (pod ServiceAccount token) is
// preserved and Impersonate-* fields are populated from the session.
func ApplyAuth(base *rest.Config, sess *Session, mode config.AuthMode) *rest.Config {
	out := rest.CopyConfig(base)
	switch mode {
	case config.AuthModeImpersonation:
		out.Impersonate = rest.ImpersonationConfig{
			UserName: sess.Username,
			Groups:   sess.Groups,
		}
	default:
		out.BearerToken = sess.IDToken
		out.BearerTokenFile = ""
	}
	return out
}
