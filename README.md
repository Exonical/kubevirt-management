# kubevirt-management

Web UI for managing [KubeVirt](https://kubevirt.io/) virtual machines.
Deploys alongside KubeVirt and uses the cluster's identity provider — no
separate user database, no app-level RBAC. If `kubectl` would let you do
it, this UI will too; if not, the kube-apiserver's 403 propagates back.

## Status

Initial scaffold. The first vertical slice (OIDC login → list namespaces →
list `VirtualMachines` in a namespace) is wired end to end. VM power
actions, templates, networking, storage, console, and metrics are tracked
in [`requirements.md`](./requirements.md).

## Architecture

```
┌────────────┐    OIDC      ┌──────────────────┐    Bearer / Impersonate    ┌────────────────┐
│  Browser   │ ─────────►  │  kubevirt-mgmt   │ ───────────────────────►  │ kube-apiserver │
│ (React 19) │             │  (Go, chi)       │                            │  + KubeVirt    │
└────────────┘  cookie     │  serves SPA      │                            └────────────────┘
                           └──────────────────┘
```

* **Backend:** Go 1.26 + `chi`, OIDC via `coreos/go-oidc`, K8s via
  `client-go` (typed for core resources, dynamic for KubeVirt CRDs
  initially). Serves the SPA from an `embed.FS`.
* **Frontend:** React 19 + Vite + TypeScript, Tailwind v3, shadcn-style
  components, TanStack Query + TanStack Router.
* **Auth:** OIDC token passthrough by default. Users authenticate via the
  cluster's IdP; the ID token is stored in an encrypted httpOnly session
  cookie and forwarded to kube-apiserver on every request. Impersonation
  mode is available for clusters that don't trust an OIDC issuer.
* **No user database.** All authorization is enforced by Kubernetes RBAC.

## Repo layout

```
cmd/server/             Entry point
internal/
  api/                  HTTP handlers (chi)
  auth/                 OIDC, session cookies, middleware, impersonation
  config/               Env-based configuration
  kube/                 Per-user kube client factory
  kubevirt/             KubeVirt CRD access (dynamic client for now)
  server/               Router wiring
  static/dist/          Embedded SPA assets (replaced at build time)
  version/              Build metadata
web/                    Vite + React 19 SPA
deploy/
  helm/kubevirt-management/   Helm chart
  dex/                  Dex config for local OIDC
.github/workflows/      CI: go build/test/vet, golangci-lint, frontend, helm lint
Dockerfile              Multi-stage: web build → go build → distroless
docker-compose.yml      Local Dex for dev
Makefile                Convenience targets
```

## Local development

### 1. Backend

```bash
# Pull deps + run a Dex OIDC provider for the login flow
docker compose up -d

# Generate dev keys + env (one-time)
eval "$(make dev-keys)"

# Make sure your kubeconfig points at a cluster with KubeVirt installed
# (e.g. `kind` + the KubeVirt operator). The dev flow uses impersonation
# so your local user must have `impersonate` on `users` and `groups`.
go run ./cmd/server
```

### 2. Frontend (separate terminal)

```bash
cd web
npm install
npm run dev  # http://localhost:5173, proxies /api to :8080
```

Open <http://localhost:5173>. Sign in with the Dex static user
`admin@example.com` / `password`.

## Deploying

A Helm chart is provided at `deploy/helm/kubevirt-management/`:

```bash
helm upgrade --install kubevirt-management deploy/helm/kubevirt-management \
  --namespace kubevirt --create-namespace \
  --set auth.mode=oidc \
  --set auth.oidc.issuer=https://your-idp.example.com \
  --set auth.oidc.clientID=kubevirt-management \
  --set auth.oidc.redirectURL=https://kubevirt-management.example.com/api/auth/callback \
  --set auth.clientSecret=$OIDC_CLIENT_SECRET \
  --set auth.sessionHashKey=$(openssl rand -hex 64) \
  --set auth.sessionBlockKey=$(openssl rand -hex 32) \
  --set ingress.enabled=true \
  --set ingress.host=kubevirt-management.example.com
```

The pod's ServiceAccount only needs `selfsubjectaccessreviews` /
`tokenreviews` permissions (and, in impersonation mode, the
`impersonate` verb on users/groups). All other operations are performed
as the authenticated user.

## License

TBD.
