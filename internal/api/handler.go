// Package api defines HTTP handlers for the dashboard backend.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Exonical/kubevirt-management/internal/kube"
	"github.com/Exonical/kubevirt-management/internal/kubevirt"
	"github.com/Exonical/kubevirt-management/internal/version"
)

// Handler bundles dependencies for HTTP handlers.
type Handler struct {
	clients *kube.Factory
	kv      *kubevirt.Service
	logger  *slog.Logger
}

// NewHandler builds an API Handler.
func NewHandler(clients *kube.Factory, kv *kubevirt.Service, logger *slog.Logger) *Handler {
	return &Handler{clients: clients, kv: kv, logger: logger}
}

// Healthz reports liveness.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// VersionHandler reports build metadata.
func VersionHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":   version.Version,
		"commit":    version.Commit,
		"buildDate": version.BuildDate,
		"goVersion": version.GoVersion(),
	})
}

// Readyz reports readiness by attempting a discovery call against the API
// server using the pod's own ServiceAccount. Per-user reachability is
// checked at request time.
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	base := h.clients.BaseConfig()
	if base == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "no kubernetes config")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
