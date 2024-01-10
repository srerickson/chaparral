package config_test

import (
	"os"
	"testing"

	"github.com/srerickson/chaparral/cmd/chap/config"
)

func TestLoadEnv(t *testing.T) {
	const testVal = "test-value"
	const envVar = "CHAPARRAL_SERVER"
	preVal, prevSet := os.LookupEnv(envVar)
	if prevSet {
		defer os.Setenv(envVar, preVal)
	}
	os.Setenv(envVar, testVal)
	cfg := config.LoadEnv()
	if cfg.ServerURL != testVal {
		t.Errorf("LoadConfig() didn't return expected value: got=%q, expect=%q", cfg.ServerURL, testVal)
	}

}
