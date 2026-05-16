# kubevirt-management — Requirements

Living document tracking what's in, out, and on deck.

## Architecture decisions (locked)

- **Go 1.26** for the backend.
- **React 19 + Vite + TypeScript** for the frontend, **served as static
  assets by the Go binary**. No Next.js (no SEO, no anonymous traffic, no
  SSR benefit).
- **OIDC token passthrough** as the primary auth mode. The cluster's
  kube-apiserver must trust the same OIDC issuer. The user's ID token is
  stored in an encrypted httpOnly session cookie and forwarded as the
  Bearer token on every K8s call. **No user database.**
- **Impersonation mode** as the fallback for clusters without
  apiserver-side OIDC. The pod's ServiceAccount has the `impersonate`
  verb and sets `Impersonate-User`/`Impersonate-Group` headers from the
  OIDC identity.
- **All authorization is enforced by Kubernetes RBAC.** The app does no
  permission checks beyond `SelfSubjectAccessReview` for cosmetics.
- Vanilla K8s. No OpenShift-specific code paths in v1.
- Dex for local dev OIDC; production users plug in their own IdP.
- Multi-cluster via kubeconfig contexts is planned but out of v1 scope.

## MVP (v1) — must ship

- [x] Repo scaffold (Go + React, single binary)
- [x] OIDC login / callback / logout / me
- [x] Encrypted session cookie (gorilla/securecookie)
- [x] OIDC token-passthrough mode
- [x] Impersonation mode
- [x] Per-user kube client factory
- [x] `GET /api/namespaces` (lists what the user can see)
- [x] `GET /api/namespaces/{ns}/virtualmachines` (list KubeVirt VMs)
- [x] `GET /api/namespaces/{ns}/virtualmachines/{name}` (get one VM)
- [x] React shell: header, sign-in, namespace browser, VM list
- [x] Multi-stage Dockerfile
- [x] Helm chart with ServiceAccount/ClusterRole/Secret/Deployment/Ingress
- [x] CI: go build/test/vet, golangci-lint, frontend lint/typecheck/build,
      helm lint
- [ ] VM power actions: start, stop, restart, pause, unpause
- [ ] VM detail page: spec YAML view, conditions, events, VMI status
- [ ] VM creation wizard from `VirtualMachineClusterInstancetype` +
      `VirtualMachineClusterPreference`
- [ ] VM creation wizard from raw YAML
- [ ] Templates: list/read/edit/delete `VirtualMachineClusterInstancetype`
      and `VirtualMachineClusterPreference`
- [ ] Networking tab per VM: list `NetworkAttachmentDefinition`s,
      show attached networks, IPs from VMI status
- [ ] Storage tab per VM: list DataVolumes/PVCs, attach/detach disks
- [ ] VNC console via WebSocket proxy to virt-api
- [ ] Serial console via WebSocket proxy
- [ ] Events panel per VM
- [ ] SelfSubjectAccessReview-driven greying of unauthorized actions
- [ ] Live updates: SSE for VM/VMI status changes

## v2 — likely

- [ ] Live migration UX with policy preview
- [ ] Snapshots & restore (VirtualMachineSnapshot CRD)
- [ ] Hotplug volumes / NICs
- [ ] Cloud-init template builder
- [ ] Multi-cluster: kubeconfig context switcher
- [ ] Prometheus integration: per-VM CPU/memory/disk panels
- [ ] Audit log of actions taken via the UI

## Explicitly out of scope (for now)

- App-level user/role/permission database — never. RBAC stays in K8s.
- Multi-tenancy beyond namespaces.
- Custom RBAC editor (use `kubectl` / GitOps).
- Backup orchestration (Velero etc.).
- Provisioning bare-metal nodes.

## Open questions to confirm with user

1. Branding / product name on the UI header? Currently "KubeVirt
   Management".
2. Hosted artifact destination — GHCR under `Exonical/kubevirt-management`?
3. License — MIT, Apache-2.0?
4. Target KubeVirt version compatibility floor (currently assume v1.0+).
