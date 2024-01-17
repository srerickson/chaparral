package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/srerickson/chaparral"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/chaparral/server/backend"
	"github.com/srerickson/chaparral/server/chapdb"

	"github.com/go-chi/httplog/v2"
	"github.com/kkyr/fig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var configFile = flag.String("c", "", "config file")

type config struct {
	// TODO: add Backends field: a map of backend IDs to Backend config. Allow
	// Roots to specify a backend by its ID. Uploader should also specify a backend.
	Backend string `fig:"backend" default:"file://."`
	Roots   []root `fig:"roots"`
	Uploads string `fig:"uploads" default:"uploads"`
	Listen  string `fig:"listen"`
	StateDB string `fig:"db" default:"chaparral.sqlite3"`
	AuthPEM string `fig:"auth_pem" default:"chaparral.pem"`
	TLSCert string `fig:"tls_cert"`
	TLSKey  string `fig:"tls_key"`
	Debug   bool   `fig:"debug"`
}

type root struct {
	ID   string `fig:"id"`
	Path string `fig:"path" validate:"required"`
	Init *struct {
		Layout      string `fig:"layout" default:"0002-flat-direct-storage-layout"`
		Description string `fig:"description"`
	} `fig:"init"`
}

var loggerOptions = httplog.Options{
	JSON:             true,
	Concise:          true,
	RequestHeaders:   true,
	MessageFieldName: "message",
}

func main() {
	flag.Parse()
	var conf config

	figOpts := []fig.Option{
		fig.UseEnv("CHAPARRAL"),
		fig.File(*configFile),
	}
	if *configFile == "" {
		// configure through environment variable only
		figOpts = append(figOpts, fig.IgnoreFile())
	}

	if err := fig.Load(&conf, figOpts...); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	if conf.Debug {
		loggerOptions.LogLevel = slog.LevelDebug
	}
	logger := httplog.NewLogger("chaparral", loggerOptions)
	backend, err := newBackend(conf.Backend)
	if err != nil {
		logger.Error(fmt.Sprintf("initializing backend: %v", err))
		os.Exit(1)
	}
	fsys, err := backend.NewFS()
	if err != nil {
		logger.Error(fmt.Sprintf("backend configuration has errors: %v", err))
		return
	}
	if ok, err := backend.IsAccessible(); !ok {
		logger.Error(fmt.Sprintf("backend not accessible: %v", err), "storage", conf.Backend)
	}
	roots := []*server.StorageRoot{}
	for _, rootConfig := range conf.Roots {
		var init *server.StorageInitializer
		if rootConfig.Init != nil {
			init = &server.StorageInitializer{
				Description: rootConfig.Init.Description,
				Layout:      rootConfig.Init.Layout,
			}
		}
		r := server.NewStorageRoot(rootConfig.ID, fsys, rootConfig.Path, init)
		roots = append(roots, r)
	}

	if conf.Uploads != "" {
		// TODO
	}
	// authentication config (load RSA key used in JWS signing)
	authKey, err := loadRSAKey(conf.AuthPEM)
	if err != nil {
		logger.Error("error loading auth keyfile", "error", err.Error())
		os.Exit(1)
	}
	db, err := chapdb.Open("sqlite3", conf.StateDB, true)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	//mgr := uploader.NewManager()

	defer db.Close()
	mux := server.New(
		server.WithStorageRoots(roots...),
		server.WithLogger(logger.Logger),
		server.WithAuthUserFunc(server.DefaultAuthUserFunc(&authKey.PublicKey)),
		server.WithAuthorizer(server.DefaultPermissions()),
		server.WithMiddleware(
			// log all requests
			httplog.RequestLogger(logger),
		),
	)
	// healthcheck endpoint
	mux.Handle("/alive", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))
	// grpc reflection endpoint
	// reflector := grpcreflect.NewStaticReflector(server.AccessServiceName, server.CommitServiceName)
	// mux.Mount(grpcreflect.NewHandlerV1(reflector))
	// mux.Mount(grpcreflect.NewHandlerV1Alpha(reflector))

	tlsCfg, err := newTLSConfig(conf.TLSCert, conf.TLSKey, "")
	if err != nil {
		logger.Error(fmt.Sprintf("in TLS config: %v", err))
		os.Exit(1)
	}
	httpSrv := http.Server{
		Addr:      conf.Listen,
		TLSConfig: tlsCfg,
	}

	logger.Info("starting server",
		"code_version", chaparral.Commit,
		"storage", conf.Backend,
		"root", conf.Roots,
		"uploads", conf.Uploads,
		"listen", conf.Listen,
		"db", conf.StateDB,
		"auth_pem", conf.AuthPEM,
		"tls_cert", conf.TLSCert,
		"tls_key", conf.TLSKey,
		"h2c", tlsCfg == nil,
	)

	// handle shutdown
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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
		httpSrv.ListenAndServeTLS("", "")
	}
	if errors.Is(http.ErrServerClosed, srvErr) {
		srvErr = nil
	}
	if srvErr != nil {
		logger.Error("server error: " + srvErr.Error())
	}
}

func newBackend(storage string) (server.Backend, error) {
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
