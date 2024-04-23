package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/chaparral/server/uploader"
)

// New returns a server mux with registered handlers for access and commit
// services.
func New(opts ...Option) *chi.Mux {
	cfg := config{}
	for _, o := range opts {
		o(&cfg)
	}
	mux := chi.NewMux()
	if cfg.logger != nil {
		mux.Use(LoggerMiddleware(cfg.logger))
	}
	if cfg.authFunc != nil {
		mux.Use(AuthUserMiddleware(cfg.authFunc))
	}
	if len(cfg.middleware) > 0 {
		mux.Use(cfg.middleware...)
	}
	mux.Mount(cfg.chaparral.CommitServiceHandler())
	mux.Mount(cfg.chaparral.AccessServiceHandler())
	mux.Mount(cfg.chaparral.ManageServiceHandler())
	return mux
}

// Option is used to configure the server mux created with New
type Option func(*config)

type config struct {
	chaparral
	middleware chi.Middlewares
	logger     *slog.Logger
	authFunc   AuthUserFunc
}

func WithStorageRoots(roots ...*store.StorageRoot) Option {
	return func(c *config) {
		c.roots = make(map[string]*store.StorageRoot, len(roots))
		for _, g := range roots {
			c.roots[g.ID()] = g
		}
	}
}

func WithUploaderManager(mgr *uploader.Manager) Option {
	return func(c *config) {
		c.chaparral.uploadMgr = mgr
	}
}

// WithAuthorizer sets the Authorizer used to determine if user are authorize
// user actions on resources.
func WithAuthorizer(auth Authorizer) Option {
	return func(c *config) {
		c.auth = auth
	}
}

// WithAuthUserFun sets the function used to resolve users from requests
func WithAuthUserFunc(fn AuthUserFunc) Option {
	return func(c *config) {
		c.authFunc = fn
	}
}

// WithLogger sets the logger that is added to all requests contexts and used by
// service hanlders.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func WithMiddleware(mids ...func(http.Handler) http.Handler) Option {
	return func(c *config) {
		c.middleware = append(c.middleware, mids...)
	}
}

func (c *chaparral) AccessServiceHandler() (string, http.Handler) {
	return (&AccessService{chaparral: c}).Handler()
}

func (c *chaparral) CommitServiceHandler() (string, http.Handler) {
	return (&CommitService{chaparral: c}).Handler()
}

func (c *chaparral) ManageServiceHandler() (string, http.Handler) {
	return (&ManageService{chaparral: c}).Handler()
}

// close any resource created with New().
func (c *chaparral) Close() error {
	return nil
}

// chaparral represents complete chaparral server state.
type chaparral struct {
	roots     map[string]*store.StorageRoot
	auth      Authorizer
	uploadMgr *uploader.Manager
}

func (c *chaparral) StorageRoot(id string) (*store.StorageRoot, error) {
	if r := c.roots[id]; r != nil {
		return r, nil
	}
	return nil, fmt.Errorf("unknown storage root: %q", id)
}

func (c *chaparral) Allowed(ctx context.Context, act, resource string) bool {
	if c.auth == nil {
		return true
	}
	return c.auth.Allowed(ctx, act, resource)
}
