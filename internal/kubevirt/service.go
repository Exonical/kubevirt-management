// Package kubevirt wraps KubeVirt CRD access. For the initial scaffold we
// use a dynamic client so we are not forced to track kubevirt.io/client-go
// version drift; we will introduce typed accessors in subsequent PRs.
package kubevirt

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/Exonical/kubevirt-management/internal/auth"
	"github.com/Exonical/kubevirt-management/internal/kube"
)

// GroupVersion of the KubeVirt API the dashboard targets.
const (
	GroupName    = "kubevirt.io"
	GroupVersion = "v1"
)

// VirtualMachineGVR identifies the VirtualMachine CRD.
var VirtualMachineGVR = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  GroupVersion,
	Resource: "virtualmachines",
}

// VirtualMachineInstanceGVR identifies the VirtualMachineInstance CRD.
var VirtualMachineInstanceGVR = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  GroupVersion,
	Resource: "virtualmachineinstances",
}

// VirtualMachineSummary is the minimal projection of a KubeVirt
// VirtualMachine returned by the API.
type VirtualMachineSummary struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	UID         string            `json:"uid"`
	Status      string            `json:"status"`
	Ready       bool              `json:"ready"`
	RunStrategy string            `json:"runStrategy,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	NodeName    string            `json:"nodeName,omitempty"`
}

// Service exposes operations against KubeVirt CRDs.
type Service struct {
	clients *kube.Factory
	logger  *slog.Logger
}

// NewService constructs a Service.
func NewService(clients *kube.Factory, logger *slog.Logger) *Service {
	return &Service{clients: clients, logger: logger}
}

// ListVirtualMachines returns VMs in a namespace as seen by the given user.
// An empty namespace means "all namespaces" (subject to RBAC).
func (s *Service) ListVirtualMachines(ctx context.Context, sess *auth.Session, namespace string) ([]VirtualMachineSummary, error) {
	dyn, err := s.clients.Dynamic(sess)
	if err != nil {
		return nil, fmt.Errorf("build dynamic client: %w", err)
	}
	unstructuredList, err := dyn.Resource(VirtualMachineGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list virtualmachines: %w", err)
	}

	out := make([]VirtualMachineSummary, 0, len(unstructuredList.Items))
	for _, item := range unstructuredList.Items {
		summary := VirtualMachineSummary{
			Namespace:   item.GetNamespace(),
			Name:        item.GetName(),
			UID:         string(item.GetUID()),
			Labels:      item.GetLabels(),
			RunStrategy: stringField(item.Object, "spec", "runStrategy"),
			Status:      stringField(item.Object, "status", "printableStatus"),
			Ready:       boolField(item.Object, "status", "ready"),
		}
		out = append(out, summary)
	}
	return out, nil
}

// GetVirtualMachine returns a single VM by name in a namespace.
func (s *Service) GetVirtualMachine(ctx context.Context, sess *auth.Session, namespace, name string) (map[string]any, error) {
	dyn, err := s.clients.Dynamic(sess)
	if err != nil {
		return nil, fmt.Errorf("build dynamic client: %w", err)
	}
	obj, err := dyn.Resource(VirtualMachineGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get virtualmachine: %w", err)
	}
	return obj.Object, nil
}

func stringField(obj map[string]any, path ...string) string {
	v, ok := traverse(obj, path...)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func boolField(obj map[string]any, path ...string) bool {
	v, ok := traverse(obj, path...)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func traverse(obj map[string]any, path ...string) (any, bool) {
	var cur any = obj
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
