package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

var uploader = &uploaderCmd{
	Command: &cobra.Command{
		Use:   "uploader",
		Short: "manage uploaders used for new object versions",
		Long:  ``,
	},
}

func init() {
	uploader.Command.Flags().BoolVar(&uploader.delete, "delete", false, "delete the uploader(s) given as arguments")
	uploader.Command.Run = ClientRunFunc(uploader)
	root.Command.AddCommand(uploader.Command)
}

type uploaderCmd struct {
	*cobra.Command
	delete bool
}

func (cmd *uploaderCmd) Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error {
	var err error
	uploaderIDs := args
	if len(uploaderIDs) == 0 {
		ups, err := cli.ListUploaders(ctx)
		if err != nil {
			return err
		}
		for _, up := range ups {
			fmt.Println(up.ID, up.Description)
		}
		return nil
	}
	for _, id := range uploaderIDs {
		up, getErr := cli.GetUploader(ctx, id)
		if getErr != nil {
			err = errors.Join(err, fmt.Errorf("uploader %q: %w", id, getErr))
			continue
		}
		if cmd.delete {
			if delErr := cli.DeleteUploader(ctx, id); delErr != nil {
				err = errors.Join(err, fmt.Errorf("uploader %q: %w", id, delErr))
			} else {
				fmt.Println("delete", id)
			}
			continue
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(up); err != nil {
			return err
		}
	}
	return err
}
