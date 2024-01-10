package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

const (
	chaparral = "chaparral"
	dirMode   = 0775
	fileMode  = 0664
)

type rootCmd struct {
	*cobra.Command
	cfgFile string
	cfg     *cfg.Config
}

// rootCmd represents the base command when called without any subcommands
var root = &rootCmd{
	Command: &cobra.Command{
		Use:   "chap",
		Short: "chaparral client",
		Long:  ``,
	},
}

func Execute() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(root.initConfig)
	root.Command.PersistentFlags().StringVar(&root.cfgFile, "config", "", "path to global config file")
}

func (root *rootCmd) initConfig() {
	conf, err := cfg.Load(root.cfgFile, "")
	if root.cfgFile == "" && errors.Is(err, fs.ErrNotExist) {
		err = nil // okay if default config file doesn't exist
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ignoring some configs due to errors:", err.Error())
	}
	root.cfg = &conf
}

type Runner interface {
	Run(ctx context.Context, conf *cfg.Config, args []string) error
}

func RunFunc[t Runner](runner t) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if err := runner.Run(ctx, root.cfg, args); err != nil {
			fmt.Fprintln(os.Stderr, cmd.Name(), "error:", err.Error())
			os.Exit(1)
		}
	}
}

type ClientRunner interface {
	Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error
}

func ClientRunFunc[t ClientRunner](runner t) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		httpcli, err := root.cfg.HttpClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, cmd.Name(), "client error:", err.Error())
			os.Exit(1)
		}
		cli := client.NewClient(httpcli, root.cfg.ServerURL(root.cfg.ServerURL("")))
		if err := runner.Run(ctx, cli, root.cfg, args); err != nil {
			fmt.Fprintln(os.Stderr, cmd.Name(), "error:", err.Error())
			os.Exit(1)
		}
	}
}
