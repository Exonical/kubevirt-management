// Package server wires the chi router, middleware, and handlers together.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/Exonical/kubevirt-management/internal/api"
	"github.com/Exonical/kubevirt-management/internal/auth"
	"github.com/Exonical/kubevirt-management/internal/config"
	"github.com/Exonical/kubevirt-management/internal/kube"
	"github.com/Exonical/kubevirt-management/internal/kubevirt"
	"github.com/Exonical/kubevirt-management/internal/static"
)

// Server bundles all dependencies needed to serve HTTP requests.
type Server struct {
	cfg     *config.Config
	logger  *slog.Logger
	router  *chi.Mux
	sess    *auth.SessionStore
	oidc    *auth.OIDCManager
	clients *kube.Factory
	kv      *kubevirt.Service
}

// New constructs a Server. It performs OIDC discovery and prepares the
// kube client factory, but does not start listening.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	sess := auth.NewSessionStore(cfg)

	oidcMgr, err := auth.NewOIDCManager(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	factory, err := kube.NewFactory(cfg, logger)
	if err != nil {
		return nil, err
	}

	kvSvc := kubevirt.NewService(factory, logger)

	s := &Server{
		cfg:     cfg,
		logger:  logger,
		sess:    sess,
		oidc:    oidcMgr,
		clients: factory,
		kv:      kvSvc,
	}
	s.router = s.buildRouter()
	return s, nil
}

// Handler returns the root http.Handler.
func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) buildRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(requestLogger(s.logger))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Compress(5))

	authMW := auth.NewMiddleware(s.sess, s.logger)
	authHandler := auth.NewHandler(s.cfg, s.oidc, s.sess, s.logger)
	apiHandler := api.NewHandler(s.clients, s.kv, s.logger)

	r.Route("/api", func(r chi.Router) {
		// NoCache only applies to API responses; the SPA static handler
		// needs to keep its long-lived cache headers for hashed assets.
		r.Use(chimw.NoCache)

		r.Get("/healthz", api.Healthz)
		r.Get("/readyz", apiHandler.Readyz)
		r.Get("/version", api.VersionHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", authHandler.Login)
			r.Get("/callback", authHandler.Callback)
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authMW.RequireAuth(authHandler.Me))
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW.Require)
			r.Get("/namespaces", apiHandler.ListNamespaces)
			r.Get("/namespaces/{ns}/virtualmachines", apiHandler.ListVirtualMachines)
			r.Get("/namespaces/{ns}/virtualmachines/{name}", apiHandler.GetVirtualMachine)
		})
	})

	// SPA: serve embedded assets, fall back to index.html for client-side routes.
	r.Handle("/*", static.Handler())

	return r
}
