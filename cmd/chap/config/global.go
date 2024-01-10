package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// map of struct fields for dynamic
var globalConfigKeys map[string]struct{} // set during init

func init() {
	globalConfigKeys = map[string]struct{}{}
	for _, f := range reflect.VisibleFields(reflect.TypeOf(GlobalConfig{})) {
		if tag, ok := f.Tag.Lookup("json"); ok {
			tag, _, _ := strings.Cut(tag, ",")
			globalConfigKeys[tag] = struct{}{}
		}
	}
}

type GlobalConfig struct {
	ServerURL      string `json:"server_url"`
	AuthToken      string `json:"auth_token"`
	StorageGroupID string `json:"storage_group_id,omitempty"`
	StorageRootID  string `json:"storage_root_id,omitempty"`
	UserName       string `json:"user_name"`
	UserEmail      string `json:"user_email"`

	ClientCrt string `json:"client_crt,omitempty"`
	ClientKey string `json:"client_key,omitempty"`
	ServerCA  string `json:"server_ca,omitempty"`

	// DefaultDigestAlgorithm sets the digest algorithm for new objects.
	DefaultDigestAlgorithm string `json:"default_digest_algorithm"`
}

func SetGlobal(cfgFile, key, val string) (err error) {
	if _, ok := globalConfigKeys[key]; !ok {
		return fmt.Errorf("unknown configuration key %q", key)
	}
	cfg := map[string]any{}
	cfgFile = GlobalConfigPath(cfgFile)
	err = decodeGlobalConfig(cfgFile, &cfg)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
		cfg = map[string]any{}
	}
	if err != nil {
		return fmt.Errorf("reading existing config: %w", err)
	}
	cfg[key] = val
	return writeGlobalConfig(cfgFile, cfg)
}

// GlobalConfig resolves that path to the global config. Its argument is an
// optional custom path. It returns an empty string if a path cannot be
// determined. Otherwise, the returned path is always an absolute path.
func GlobalConfigPath(custom string) string {
	if custom != "" {
		if abs, err := filepath.Abs(custom); err == nil {
			return abs
		}
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cfgDir, chaparral, configFile)
}

func getGlobalConfig(custom string) (cfg GlobalConfig, err error) {
	err = decodeGlobalConfig(custom, &cfg)
	if err != nil {
		cfg = GlobalConfig{}
	}
	return
}

func decodeGlobalConfig(name string, v any) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func writeGlobalConfig(cfgPath string, cfg any) error {
	absPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), dirMode); err != nil {
		return err
	}
	writer, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("writing new configuration: %w", err)
	}
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "\t")
	return enc.Encode(cfg)
}
