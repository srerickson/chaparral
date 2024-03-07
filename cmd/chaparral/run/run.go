package run

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/httplog/v2"
	"github.com/srerickson/chaparral"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/chaparral/server/backend"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const healthCheck = "/alive"

type Config struct {
	Backend     string                 `fig:"backend" default:"file://."`
	Roots       []Root                 `fig:"roots"`
	Uploads     string                 `fig:"uploads"`
	Listen      string                 `fig:"listen"`
	StateDB     string                 `fig:"db" default:"chaparral.sqlite3"`
	AuthPEM     string                 `fig:"auth_pem" default:"chaparral.pem"`
	TLSCert     string                 `fig:"tls_cert"`
	TLSKey      string                 `fig:"tls_key"`
	Debug       bool                   `fig:"debug"`
	Permissions server.RolePermissions `fig:"permissions"`
}

type Root struct {
	ID   string `fig:"id"`
	Path string `fig:"path" validate:"required"`
	Init *struct {
		Layout      string `fig:"layout" default:"0002-flat-direct-storage-layout"`
		Description string `fig:"description"`
	} `fig:"init"`
}

func Run(ctx context.Context, conf *Config) error {

	var loggerOptions = httplog.Options{
		JSON:             true,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	}

	if conf.Debug {
		loggerOptions.LogLevel = slog.LevelDebug
	}
	logger := httplog.NewLogger("chaparral", loggerOptions)

	logger.Debug("chaparral version",
		"code_version", chaparral.CODE_VERSION,
		"version", chaparral.VERSION)

	// sqlite3 for server state
	logger.Debug("opening database...", "config", conf.StateDB)
	db, err := chapdb.Open("sqlite3", conf.StateDB, true)
	if err != nil {
		return fmt.Errorf("loading database: %w", err)
	}
	defer db.Close()
	chapDB := (*chapdb.SQLiteDB)(db)

	logger.Debug("initializing backend...", "config", conf.Backend)
	backend, err := newBack(conf.Backend)
	if err != nil {
		return fmt.Errorf("initializing backend: %w", err)
	}

	fsys, err := backend.NewFS()
	if err != nil {
		return fmt.Errorf("backend configuration has errors: %w", err)
	}
	if ok, err := backend.IsAccessible(); !ok {
		return fmt.Errorf("backend is not accessible: %w", err)
	}
	var rootPaths []string
	var roots []*store.StorageRoot
	for _, rootConfig := range conf.Roots {
		var init *store.StorageRootInitializer
		if rootConfig.Init != nil {
			init = &store.StorageRootInitializer{
				Description: rootConfig.Init.Description,
				Layout:      rootConfig.Init.Layout,
			}
		}
		logger.Debug("using storage root",
			"id", rootConfig.ID,
			"path", rootConfig.Path,
			"initialize", init != nil)
		r := store.NewStorageRoot(rootConfig.ID, fsys, rootConfig.Path, init, chapDB)
		roots = append(roots, r)
		rootPaths = append(rootPaths, rootConfig.Path)
	}

	// authentication config (load RSA key used in JWS signing)
	logger.Debug("using keyfile...", "config", conf.AuthPEM)
	authKey, err := loadRSAKey(conf.AuthPEM)
	if err != nil {
		return fmt.Errorf("loading auth keyfile: %w", err)
	}

	// upload manager is required for allowing uploads
	var mgr *uploader.Manager
	if conf.Uploads != "" {
		logger.Debug("uploads are enabled", "config", conf.Uploads)
		mgr = uploader.NewManager(fsys, conf.Uploads, chapDB)
		rootPaths = append(rootPaths, conf.Uploads)
	}

	if pathConflict(rootPaths...) {
		return fmt.Errorf("storage root and uploader paths have conflicts: %s", strings.Join(rootPaths, ", "))
	}

	// role definitions
	roles := conf.Permissions
	if roles.Empty() {
		// allow everything:
		roles.Default = server.Permissions{"*": []string{"*"}}
	}
	logger.Debug("default role", "permissions", roles.Default)

	mux := server.New(
		server.WithStorageRoots(roots...),
		server.WithUploaderManager(mgr),
		server.WithLogger(logger.Logger),
		server.WithAuthUserFunc(server.JWSAuthFunc(&authKey.PublicKey)),
		server.WithAuthorizer(roles),
		server.WithMiddleware(
			// log all requests
			httplog.RequestLogger(logger, []string{healthCheck}),
		),
	)
	// healthcheck endpoint
	mux.Handle(healthCheck, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))

	// grpc reflection endpoint
	// reflector := grpcreflect.NewStaticReflector(server.AccessServiceName, server.CommitServiceName)
	// mux.Mount(grpcreflect.NewHandlerV1(reflector))
	// mux.Mount(grpcreflect.NewHandlerV1Alpha(reflector))

	logger.Debug("TLS config", "cert", conf.TLSCert, "key", conf.TLSKey)
	tlsCfg, err := newTLSConfig(conf.TLSCert, conf.TLSKey, "")
	if err != nil {
		return fmt.Errorf("TLS config errors: %w", err)
	}
	httpSrv := http.Server{
		Addr:      conf.Listen,
		TLSConfig: tlsCfg,
	}

	logger.Info("starting server", "listen", conf.Listen, "h2c", tlsCfg == nil)

	// handle shutdown
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		logger.Info("shutting down ...", "deadline", "2 mins")
		if err := httpSrv.Shutdown(ctx); err != nil {
			if errors.Is(context.DeadlineExceeded, err) {
				logger.Error("server didn't shutdown gracefully")
				httpSrv.Close()
			}
		}
		if err := db.Close(); err != nil {
			logger.Error("shutting down database: " + err.Error())
		}
	}()

	var srvErr error
	switch {
	case httpSrv.TLSConfig == nil:
		httpSrv.Handler = h2c.NewHandler(mux, &http2.Server{})
		srvErr = httpSrv.ListenAndServe()
	default:
		httpSrv.Handler = mux
		srvErr = httpSrv.ListenAndServeTLS("", "")
	}
	if errors.Is(http.ErrServerClosed, srvErr) {
		srvErr = nil
	}
	return srvErr
}

type back interface {
	Name() string
	IsAccessible() (bool, error)
	NewFS() (ocfl.WriteFS, error)
}

func newBack(storage string) (back, error) {
	kind, loc, _ := strings.Cut(storage, "://")
	switch kind {
	case "file":
		return &backend.FileBackend{Path: loc}, nil
	case "s3":
		bucket, query, _ := strings.Cut(loc, "?")
		back := &backend.S3Backend{Bucket: bucket}
		if query != "" {
			opts, err := url.ParseQuery(query)
			if err != nil {
				return nil, err
			}
			back.Options = opts
		}
		return back, nil
	default:
		return nil, fmt.Errorf("invalid storage backen: %q", storage)
	}
}

func loadRSAKey(name string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := genRSAKey(name); err != nil {
			return nil, err
		}
		bytes, err = os.ReadFile(name)
		if err != nil {
			return nil, err
		}
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("key is not PEM encoded")
	}
	anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	switch k := anyKey.(type) {
	case *rsa.PrivateKey:
		return k, nil
	default:
		return nil, errors.New("not an rsa key")
	}
}

func genRSAKey(name string) error {
	slog.Info("generating new RSA key", "name", name)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func newTLSConfig(crt, key, clientCA string) (*tls.Config, error) {
	var tlsCfg *tls.Config
	if crt != "" || key != "" {
		var err error
		tlsCfg = &tls.Config{Certificates: make([]tls.Certificate, 1)}
		tlsCfg.Certificates[0], err = tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return nil, err
		}
		if clientCA != "" {
			pem, err := os.ReadFile(clientCA)
			if err != nil {
				return nil, err
			}
			clientCAs := x509.NewCertPool()
			if !clientCAs.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("%q not added to certool", clientCA)
			}
			tlsCfg.ClientCAs = clientCAs
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}
	return tlsCfg, nil
}

func pathConflict(paths ...string) bool {
	for i, a := range paths {
		for _, b := range paths[i+1:] {
			if a == b {
				return true
			}
			if a == "." || b == "." || strings.HasPrefix(a, b) || strings.HasPrefix(b, a) {
				return true
			}
		}
	}
	return false
}
