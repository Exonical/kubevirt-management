package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Exonical/kubevirt-management/internal/auth"
)

// ListVirtualMachines returns VMs in a namespace.
func (h *Handler) ListVirtualMachines(w http.ResponseWriter, r *http.Request) {
	sess := auth.SessionFromContext(r.Context())
	if sess == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	ns := chi.URLParam(r, "ns")
	vms, err := h.kv.ListVirtualMachines(r.Context(), sess, ns)
	if err != nil {
		writeKubeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": vms})
}

// GetVirtualMachine returns a single VM.
func (h *Handler) GetVirtualMachine(w http.ResponseWriter, r *http.Request) {
	sess := auth.SessionFromContext(r.Context())
	if sess == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	ns := chi.URLParam(r, "ns")
	name := chi.URLParam(r, "name")
	vm, err := h.kv.GetVirtualMachine(r.Context(), sess, ns, name)
	if err != nil {
		writeKubeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vm)
}
