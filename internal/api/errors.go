package api

import (
	"errors"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

// writeKubeError translates a Kubernetes API error into an HTTP response.
// Forbidden / Unauthorized / NotFound are mapped directly; everything else
// is reported as 502.
func writeKubeError(w http.ResponseWriter, err error) {
	if err == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unknown"})
		return
	}
	var statusErr *kerrors.StatusError
	if errors.As(err, &statusErr) {
		writeJSON(w, int(statusErr.ErrStatus.Code), map[string]any{
			"error":  statusErr.ErrStatus.Message,
			"reason": string(statusErr.ErrStatus.Reason),
		})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
}
