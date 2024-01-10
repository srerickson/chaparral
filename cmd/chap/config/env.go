package config

import (
	"os"
	"reflect"
)

type EnvConfig struct {
	ServerURL string `env:"CHAPARRAL_SERVER"`
}

func LoadEnv() (e EnvConfig) {
	eVal := reflect.ValueOf(&e)
	for i, f := range reflect.VisibleFields(reflect.TypeOf(e)) {
		envTag := f.Tag.Get("env")
		if envTag == "" {
			continue
		}
		val := os.Getenv(envTag)
		if val == "" {
			continue
		}
		eVal.Elem().Field(i).SetString(val) //.Field(i).Addr().SetString(val)
	}
	return
}
