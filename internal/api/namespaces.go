package api

import (
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Exonical/kubevirt-management/internal/auth"
)

// ListNamespaces returns the namespaces visible to the authenticated user.
// RBAC is enforced server-side by the kube-apiserver; we just forward the
// list request as the user.
func (h *Handler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	sess := auth.SessionFromContext(r.Context())
	if sess == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	cs, err := h.clients.Kubernetes(sess)
	if err != nil {
		h.logger.Error("build kube client", "err", err)
		writeJSONError(w, http.StatusInternalServerError, "kube client unavailable")
		return
	}
	nsList, err := cs.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeKubeError(w, err)
		return
	}
	names := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": names})
}
