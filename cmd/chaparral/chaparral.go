package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/srerickson/chaparral/cmd/chaparral/run"

	"github.com/kkyr/fig"
)

var configFile = flag.String("c", "", "config file")

func main() {
	ctx := context.Background()
	flag.Parse()
	var conf run.Config
	figOpts := []fig.Option{
		fig.UseStrict(),
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
	if err := run.Run(ctx, &conf); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
