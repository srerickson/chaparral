package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

var config = &configCmd{
	Command: &cobra.Command{
		Use:   "config",
		Short: "get and set chap configuration settings",
		Long:  ``,
	},
}

func init() {
	config.Command.Flags().StringVar(&config.setFlag, "set", "", "set config")
	config.Command.Run = RunFunc(config)
	root.Command.AddCommand(config.Command)
}

type configCmd struct {
	*cobra.Command
	setFlag string
}

func (config *configCmd) Run(ctx context.Context, conf *cfg.Config, args []string) error {
	if config.setFlag != "" {
		if len(args) == 0 {
			return fmt.Errorf("missing required argument: value for %q", config.setFlag)
		}
		if err := cfg.SetGlobal(root.cfgFile, config.setFlag, args[0]); err != nil {
			return err
		}
		// reload config
		newConf, err := cfg.Load(root.cfgFile, "")
		if err != nil {
			return err
		}
		*conf = newConf
		return nil
	}

	// print config as json
	if f := conf.GlobalConfigPath(); f != "" {
		fmt.Println("global config from:", f)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(conf.Global); err != nil {
		return err
	}
	proj, err := cfg.SearchProject("")
	if err != nil && !errors.Is(err, cfg.ErrNotProject) {
		return err
	}
	if err == nil {
		fmt.Println("local config for:", proj.Path())
		if err := enc.Encode(proj); err != nil {
			return err
		}
	}
	return nil
}
