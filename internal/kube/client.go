// Package kube builds per-user Kubernetes clients on top of a shared base
// rest.Config. The base config is the pod's in-cluster config (or a
// supplied kubeconfig for local dev); per-request clients are derived from
// it by either overriding the bearer token (OIDC mode) or setting
// Impersonate-* fields (impersonation mode).
package kube

import (
	"fmt"
	"log/slog"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Exonical/kubevirt-management/internal/auth"
	"github.com/Exonical/kubevirt-management/internal/config"
)

// Factory produces user-scoped Kubernetes clients.
type Factory struct {
	cfg    *config.Config
	base   *rest.Config
	mode   config.AuthMode
	logger *slog.Logger
}

// NewFactory loads the base rest.Config (in-cluster first, then kubeconfig
// fallback) and returns a Factory.
func NewFactory(cfg *config.Config, logger *slog.Logger) (*Factory, error) {
	base, err := loadBaseConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Factory{cfg: cfg, base: base, mode: cfg.AuthMode, logger: logger}, nil
}

// BaseConfig returns a deep copy of the base rest.Config.
func (f *Factory) BaseConfig() *rest.Config {
	return rest.CopyConfig(f.base)
}

// ForSession returns a *rest.Config configured to act as the given user.
func (f *Factory) ForSession(sess *auth.Session) *rest.Config {
	return auth.ApplyAuth(f.base, sess, f.mode)
}

// Kubernetes returns a typed clientset acting as the given user.
func (f *Factory) Kubernetes(sess *auth.Session) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(f.ForSession(sess))
}

// Dynamic returns a dynamic client acting as the given user. Used for
// KubeVirt CRDs until we wire in the typed kubevirt client.
func (f *Factory) Dynamic(sess *auth.Session) (dynamic.Interface, error) {
	return dynamic.NewForConfig(f.ForSession(sess))
}

func loadBaseConfig(cfg *config.Config) (*rest.Config, error) {
	if cfg.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(cfg.APIServerOverride, cfg.Kubeconfig)
	}
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		c, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("in-cluster config: %w", err)
		}
		applyOverrides(c, cfg)
		return c, nil
	}
	// Fall back to default kubeconfig path so local `go run` works.
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if cfg.APIServerOverride != "" {
		overrides.ClusterInfo.Server = cfg.APIServerOverride
	}
	c, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	applyOverrides(c, cfg)
	return c, nil
}

func applyOverrides(c *rest.Config, cfg *config.Config) {
	if cfg.InsecureSkipVerify {
		c.TLSClientConfig.Insecure = true
		c.TLSClientConfig.CAData = nil
		c.TLSClientConfig.CAFile = ""
	}
	if cfg.CABundlePath != "" {
		c.TLSClientConfig.CAFile = cfg.CABundlePath
	}
	c.UserAgent = "kubevirt-management/0.0.0"
}
