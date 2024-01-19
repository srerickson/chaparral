package cmd

import (
	"context"

	"github.com/spf13/cobra"
	chap "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

var del = &deleteCmd{
	Command: &cobra.Command{
		Use:   "delete",
		Short: "delete an OCFL object",
		Long:  ``,
	},
}

func init() {
	del.Command.Run = ClientRunFunc(del)
	root.Command.AddCommand(del.Command)
}

type deleteCmd struct {
	*cobra.Command
}

func (ls *deleteCmd) Run(ctx context.Context, cli *chap.Client, conf *cfg.Config, args []string) error {
	var objectID string
	if len(args) > 0 {
		objectID = args[0]
	}
	return cli.DeleteObject(ctx, conf.StorageRootID(""), objectID)
}
