package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	client "github.com/srerickson/chaparral/client"
)

const (
	configFile    = "config.json"
	chaparral     = "chaparral"
	dirMode       = 0775
	fileMode      = 0664
	defaultServer = "http://localhost:8080"
	defaultAlg    = "sha512"
	defaultSpec   = "v1.1"
)

type Config struct {
	// global (user-level) configs($HOME/.config/chaparral/config.json)
	Global GlobalConfig

	// settings from environment variables
	Env EnvConfig

	// absolute path to global config (file may not exist)
	globaPath string
}

// Load returns the current configuration settings. Its arguments are optional
// paths to global and project configuration files.
func Load(global, proj string) (c Config, err error) {
	globalPath := GlobalConfigPath(global)
	globCfg, err := getGlobalConfig(globalPath)
	if err != nil {
		return
	}
	c.Env = LoadEnv()
	c.Global = globCfg
	c.globaPath = globalPath
	return
}

func (c Config) GlobalConfigPath() string { return c.globaPath }

func (c Config) ServerURL(custom string) string {
	return First(custom, c.Env.ServerURL, c.Global.ServerURL, defaultServer)
}

func (c Config) AuthToken(custom string) string {
	return First(custom, c.Global.AuthToken)
}

func (c Config) ClientCrt(custom string) string {
	return First(custom, c.Global.ClientCrt)
}

func (c Config) ClientKey(custom string) string {
	return First(custom, c.Global.ClientKey)
}

func (c Config) ServerCA(custom string) string {
	return First(custom, c.Global.ServerCA)
}

func (c Config) UserName(custom string) string {
	return First(custom, c.Global.UserName)
}

func (c Config) UserEmail(custom string) string {
	return First(custom, c.Global.UserEmail)
}

func (c Config) StorageRootID(custom string) string {
	return First(custom, c.Global.StorageRootID)
}

func (c Config) DigestAlgorithm(alg string) string {
	return First(alg, c.Global.DefaultDigestAlgorithm, defaultAlg)
}

func (c Config) HttpClient() (*http.Client, error) {
	var (
		crt       = c.ClientCrt("")
		key       = c.ClientKey("")
		ca        = c.ServerCA("")
		authToken = c.AuthToken("")
		tlsCfg    *tls.Config
		cli       *http.Client = &http.Client{}
	)
	if crt != "" && key != "" {
		cert, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return nil, err
		}
		tlsCfg = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	if ca != "" {
		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		}
		pem, err := os.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = x509.NewCertPool()
		tlsCfg.RootCAs.AppendCertsFromPEM(pem)
	}
	if tlsCfg != nil {
		cli.Transport = &http.Transport{TLSClientConfig: tlsCfg}
	}
	if authToken != "" {
		clientWithBearerToken(cli, authToken)
	}
	return cli, nil
}

func (c Config) Client() (*client.Client, error) {
	httpCli, err := c.HttpClient()
	if err != nil {
		return nil, err
	}
	return client.NewClient(httpCli, c.ServerURL("")), nil
}

func clientWithBearerToken(cli *http.Client, token string) {
	base := cli.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cli.Transport = &bearerTokenTransport{
		Token: token,
		Base:  base,
	}
}

type bearerTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}

// returns the first non-empty value
func First[T comparable](vals ...T) T {
	var empty T
	for _, v := range vals {
		if v != empty {
			return v
		}
	}
	return empty
}
