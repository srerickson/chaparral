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
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const healthCheck = "/alive"

type Config struct {
	Backend     string                 `fig:"backend" default:"file://."`
	Roots       []Root                 `fig:"roots"`
	Uploads     string                 `fig:"uploads"`
	Listen      string                 `fig:"listen" default:":8080"`
	DB          string                 `fig:"db" default:"/tmp/chaparral.sqlite3"`
	PubkeyFile  string                 `fig:"pubkey_file"`
	Pubkey      string                 `fig:"pubkey"`
	AutoCert    *AutoCertConfig        `fix:"autocert"`
	TLSCert     string                 `fig:"tls_cert"`
	TLSKey      string                 `fig:"tls_key"`
	Debug       bool                   `fig:"debug"`
	Permissions server.RolePermissions `fig:"permissions"`
}

func (c *Config) tlsConfig() (*tls.Config, error) {
	switch {
	case c.AutoCert != nil:
		if err := os.MkdirAll(c.AutoCert.Dir, 0700); err != nil {
			return nil, err
		}
		manager := autocert.Manager{
			Cache:      autocert.DirCache(c.AutoCert.Dir),
			Email:      c.AutoCert.Email,
			HostPolicy: autocert.HostWhitelist(c.AutoCert.Domain),
		}
		return manager.TLSConfig(), nil
	case c.TLSCert != "" && c.TLSKey != "":
		cert, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
	default:
		return nil, nil
	}
}

type AutoCertConfig struct {
	Domain string
	Email  string
	// Directory for storing certificates
	Dir string `fig:"dir" default:"~/.cache/golang-autocert"`
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
	var serviceOptions []server.Option

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
	serviceOptions = append(serviceOptions,
		server.WithLogger(logger.Logger),
		server.WithMiddleware(httplog.RequestLogger(logger, []string{healthCheck})))

	logger.Debug("chaparral version",
		"code_version", chaparral.CODE_VERSION,
		"version", chaparral.VERSION)

	// sqlite3 for server state
	logger.Debug("opening database...", "config", conf.DB)
	db, err := chapdb.Open("sqlite3", conf.DB, true)
	if err != nil {
		return fmt.Errorf("loading database: %w", err)
	}
	defer db.Close()
	chapDB := (*chapdb.SQLiteDB)(db)

	logger.Debug("initializing backend...", "config", conf.Backend)
	fsys, err := newBackend(conf.Backend, logger.Logger)
	if err != nil {
		return err
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
	if len(roots) < 1 {
		logger.Warn("no storage roots configured")
	}

	serviceOptions = append(serviceOptions, server.WithStorageRoots(roots...))

	// upload manager is required for allowing uploads
	if conf.Uploads != "" {
		mgr := uploader.NewManager(fsys, conf.Uploads, chapDB)
		rootPaths = append(rootPaths, conf.Uploads)
		serviceOptions = append(serviceOptions, server.WithUploaderManager(mgr))
		logger.Debug("uploads are enabled", "config", conf.Uploads)
	}

	if pathConflict(rootPaths...) {
		return fmt.Errorf("storage root and uploader paths have conflicts: %s", strings.Join(rootPaths, ", "))
	}

	// authentication config (load RSA key used in JWS signing)
	if conf.Pubkey != "" || conf.PubkeyFile != "" {
		pubkey, err := getPubkey([]byte(conf.Pubkey), conf.PubkeyFile)
		if err != nil {
			return fmt.Errorf("parsing public key: %w", err)
		}
		authFunc, err := server.JWSAuthFunc(pubkey)
		if err != nil {
			return err
		}
		logger.Debug("using public key for JWS verification")
		serviceOptions = append(serviceOptions, server.WithAuthUserFunc(authFunc))
	}

	// role definitions
	roles := conf.Permissions
	if conf.Permissions.Empty() {
		// allow everything:
		roles.Default = server.Permissions{"*": []string{"*::*"}}
	}
	logger.Debug("applying permissions", "roles", roles)
	serviceOptions = append(serviceOptions, server.WithAuthorizer(roles))

	mux := server.New(serviceOptions...)
	// healthcheck endpoint
	mux.Handle(healthCheck, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))

	// grpc reflection endpoint
	// reflector := grpcreflect.NewStaticReflector(server.AccessServiceName, server.CommitServiceName)
	// mux.Mount(grpcreflect.NewHandlerV1(reflector))
	// mux.Mount(grpcreflect.NewHandlerV1Alpha(reflector))

	httpSrv := http.Server{Addr: conf.Listen}
	if httpSrv.TLSConfig, err = conf.tlsConfig(); err != nil {
		return fmt.Errorf("TLS config errors: %w", err)
	}

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
	case httpSrv.TLSConfig != nil:
		logger.Info("starting server", "listen", conf.Listen, "tls", true)
		httpSrv.Handler = mux
		srvErr = httpSrv.ListenAndServeTLS("", "")
	default:
		logger.Info("starting server", "listen", conf.Listen, "tls", false)
		httpSrv.Handler = h2c.NewHandler(mux, &http2.Server{})
		srvErr = httpSrv.ListenAndServe()
	}
	if errors.Is(http.ErrServerClosed, srvErr) {
		srvErr = nil
	}
	return srvErr
}

func newBackend(storage string, logger *slog.Logger) (ocfl.WriteFS, error) {
	var b interface {
		IsAccessible() (bool, error)
		NewFS() (ocfl.WriteFS, error)
	}
	kind, configStr, _ := strings.Cut(storage, "://")
	switch kind {
	case "file":
		b = &backend.FileBackend{Path: configStr}
	case "s3":
		bucket, query, _ := strings.Cut(configStr, "?")
		s3back := &backend.S3Backend{
			Bucket: bucket,
			Logger: logger,
		}
		if query != "" {
			opts, err := url.ParseQuery(query)
			if err != nil {
				return nil, fmt.Errorf("parsing s3 backend options: %w", err)
			}
			s3back.Options = opts
		}
		b = s3back
	default:
		return nil, fmt.Errorf("invalid storage backend: %q", storage)
	}
	fsys, err := b.NewFS()
	if err != nil {
		return nil, fmt.Errorf("while initializing backend: %w", err)
	}
	if ok, err := b.IsAccessible(); !ok {
		return nil, fmt.Errorf("backend is not accessible: %w", err)
	}
	return fsys, nil
}

func getPubkey(pemBytes []byte, pemFile string) (any, error) {
	if len(pemBytes) < 1 {
		var err error
		pemBytes, err = os.ReadFile(pemFile)
		if err != nil {
			return nil, err
		}
	}
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}
	return x509.ParsePKIXPublicKey(block.Bytes)
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

func pathConflict(paths ...string) bool {
	for i, a := range paths {
		for _, b := range paths[i+1:] {
			if a == b {
				return true
			}
			if a == "." || b == "." || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/") {
				return true
			}
		}
	}
	return false
}
