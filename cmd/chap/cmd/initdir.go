package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

var initdir = &initdirCmd{
	Command: &cobra.Command{
		Use:   "init object-id dir",
		Short: "initialize a local directory to push to and pull from a remote object",
		Long:  ``,
	},
}

func init() {
	initdir.Command.Run = RunFunc(initdir)
	initdir.Command.Flags().StringVar(&initdir.groupID, "group", "", "non-default storge group ID for remote object")
	initdir.Command.Flags().StringVar(&initdir.rootID, "root", "", "non-default OCFL root ID for remote object")
	root.Command.AddCommand(initdir.Command)
}

type initdirCmd struct {
	*cobra.Command
	groupID string
	rootID  string
}

func (cmd *initdirCmd) Run(ctx context.Context, conf *cfg.Config, args []string) error {
	dir := "."
	objectID := ""
	if len(args) > 0 {
		objectID = args[0]
	}
	if len(args) > 1 {
		dir = args[1]
	}
	if objectID == "" {
		return errors.New("missing required argument: object ID")
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	// load storage from global, if that's where it's set
	proj := cfg.Project{
		ObjectID:      objectID,
		StorageRootID: conf.StorageRootID(cmd.rootID),
	}
	if err := cfg.InitProject(dir, proj); err != nil {
		return err
	}
	fmt.Println("intitialized", dir, "as a local source directory for", objectID)
	return nil
}
