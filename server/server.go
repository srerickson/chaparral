package server

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/uploader"
)

// chaparral represents complete chaparral server state.
type chaparral struct {
	storeGrps map[string]*StorageGroup
	auth      Authorizer
	uploadMgr *uploader.Manager
}

type config struct {
	chaparral
	uploadRoots []uploader.Root
	db          *chapdb.SQLiteDB
	middleware  chi.Middlewares
	logger      *slog.Logger
	authFunc    AuthUserFunc
}

// New returns a server mux with registered handlers for access and commit
// services.
func New(opts ...Option) *chi.Mux {
	cfg := config{}
	for _, o := range opts {
		o(&cfg)
	}
	if len(cfg.uploadRoots) > 0 {
		cfg.chaparral.uploadMgr = uploader.NewManager(cfg.uploadRoots, cfg.db)
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
	return mux
}

// Option is used to configure the server mux created with New
type Option func(*config)

func WithStorageGroups(groups ...*StorageGroup) Option {
	return func(c *config) {
		c.storeGrps = make(map[string]*StorageGroup, len(groups))
		c.uploadRoots = []uploader.Root{}
		for _, g := range groups {
			c.storeGrps[g.ID()] = g
			if uproot := g.UploadRoot(); uproot != nil {
				c.uploadRoots = append(c.uploadRoots, *uproot)
			}
		}
	}
}

func WithSQLDB(db *sql.DB) Option {
	return func(c *config) {
		c.db = (*chapdb.SQLiteDB)(db)
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

// close any resource created with New().
func (c *chaparral) Close() error {
	return nil
}

func (c *chaparral) storageRoot(groupID, rootID string) (*StorageRoot, error) {
	group := c.storeGrps[groupID]
	if group == nil {
		return nil, fmt.Errorf("storage root %q in group %q: %w", rootID, groupID, ErrStorageRootNotFound)
	}
	root, err := group.StorageRoot(rootID)
	if err != nil {
		return nil, fmt.Errorf("storage root %q in group %q: %w", rootID, groupID, err)
	}
	return root, nil
}
