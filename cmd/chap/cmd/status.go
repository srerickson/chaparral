package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
	"github.com/srerickson/chaparral/cmd/chap/ui"
	df "github.com/srerickson/chaparral/internal/diff"
)

var status = &statusCmd{
	Command: &cobra.Command{
		Use:   "status",
		Short: "show state of the working directory",
		Long:  ``,
	},
}

func init() {
	status.Command.Run = ClientRunFunc(status)
	root.Command.AddCommand(status.Command)
}

type statusCmd struct {
	*cobra.Command
}

func (diff *statusCmd) Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error {
	proj, err := cfg.SearchProject("")
	if err != nil {
		if errors.Is(err, cfg.ErrNotProject) {
			fmt.Fprintln(os.Stderr, "Use `init` to initialize the directory")
			return cfg.ErrNotProject
		}
		return err
	}
	var (
		dir      = proj.Path()
		objectID = proj.ObjectID
		group    = proj.StorageGroupID
		root     = proj.StorageRootID
	)
	ui.PrintValues(
		"local object version", fmt.Sprintf("%s (%d)", proj.ObjectID, proj.Version),
	)

	head, err := cli.GetObjectState(ctx, group, root, objectID, 0)
	if err != nil {
		if client.IsNotFound(err) {
			fmt.Fprintln(os.Stderr, "Use `push` to upload new object version.")
			err = fmt.Errorf("%s not found", objectID)
		}
		return err
	}

	ui.PrintValues("upstream version", strconv.Itoa(int(head.Version)))
	stage, err := ui.RunStageDir(dir, head.DigestAlgorithm, client.StageSkipDirRE(chaparralRE))
	if err != nil {
		return err
	}
	result, err := df.Diff(head.State, stage.State)
	if err != nil {
		return err
	}
	if result.Empty() {
		fmt.Println("no changes")
		return nil
	}
	fmt.Println("Changes between upstream object and local source directory:")
	fmt.Println(ui.PrettyDiff(result))
	return nil
}
