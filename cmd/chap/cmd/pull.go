package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
	"github.com/srerickson/chaparral/cmd/chap/ui"
)

var pull = &pullCmd{
	Command: &cobra.Command{
		Use:   "pull [--root storage-root] [--id object-id] [--replace | -r] [dir]",
		Short: "download object versions to a local directory",
		Long:  ``,
	},
}

func init() {
	pull.Command.Flags().StringVar(&pull.objectID, "id", "", "id of object to pull (required if dir is uninitialized)")
	pull.Command.Flags().BoolVarP(&pull.replace, "replace", "r", false, "replace existing files that have changed")
	pull.Command.Flags().IntVar(&pull.vNum, "ver", 0, "used to specify object version to pull")
	pull.Command.Run = ClientRunFunc(pull)
	root.Command.AddCommand(pull.Command)
}

type pullCmd struct {
	*cobra.Command
	objectID string
	replace  bool //
	vNum     int
}

func (pull *pullCmd) Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error {
	var (
		objectID string
		groupID  string
		rootID   string
		dstDir   string
		proj     cfg.Project
		err      error
	)
	switch {
	case len(args) > 0:
		dstDir = args[0]
		proj, err = cfg.GetProject(args[0])
	default:
		proj, err = cfg.SearchProject("")
	}
	if err != nil && !errors.Is(err, cfg.ErrNotProject) {
		return err
	}
	err = nil
	objectID = cfg.First(pull.objectID, proj.ObjectID)
	groupID = conf.StorageGroupID(proj.StorageGroupID)
	rootID = conf.StorageRootID(proj.StorageRootID)
	dstDir = cfg.First(dstDir, proj.Path())
	if dstDir == "" {
		return errors.New("could not determine the destination directory for pulled content")
	}
	if objectID == "" {
		return errors.New("an object id must be set in the destination directory's config or with the '--id' flag")
	}
	if !proj.Empty() {
		ui.PrintValues("local object version", fmt.Sprintf("%s (%d)", proj.ObjectID, proj.Version))
	}
	state, err := pull.pullState(ctx, cli, groupID, rootID, objectID, dstDir)
	if err != nil {
		return err
	}

	// update to project's version to match pulled state.
	if !proj.Empty() && proj.Version != state.Version {
		proj.Version = state.Version
		if err := proj.Save(); err != nil {
			return fmt.Errorf("updating local config: %w", err)
		}
	}
	return nil
}

func (pull *pullCmd) pullState(ctx context.Context, cli *client.Client, groupID, rootID, objectID, dst string) (*client.ObjectState, error) {
	remote, err := cli.GetObjectState(ctx, groupID, rootID, objectID, pull.vNum)
	if err != nil {
		return nil, fmt.Errorf("getting object state: %w", err)
	}
	dstDir, err := filepath.Abs(dst)
	if err != nil {
		return nil, fmt.Errorf("invalid output directory name %q: %w", dst, err)
	}
	var localPathDigests map[string]string // pathmap
	exists, err := isExistingDir(dstDir)
	if err != nil {
		// can't tell if dstDir exists or not.
		return nil, err
	}
	switch exists {
	case true:
		stage, err := ui.RunStageDir(dstDir, remote.DigestAlgorithm, client.StageSkipDirRE(chaparralRE))
		if err != nil {
			return nil, err
		}
		localPathDigests = stage.State
		if pull.replace {
			break // ignore this check if it's ok to replace existing files
		}
	case false:
		// if output directory doesn't exist, it's parent directory must exist
		parentExist, err := isExistingDir(filepath.Dir(dstDir))
		if err != nil {
			return nil, err
		}
		if !parentExist {
			return nil, fmt.Errorf("parent directory for %q doesn't exist", dstDir)
		}
	}

	// dst to source
	files := map[string]string{}
	for fName, digest := range remote.State {
		if localPathDigests[fName] == digest {
			continue
		}
		dstPath := filepath.Join(dstDir, filepath.FromSlash(fName))
		files[dstPath] = digest
	}
	if err := ui.RunDownload(cli, groupID, rootID, objectID, files, pull.replace); err != nil {
		return nil, err
	}
	return remote, nil
}

// checkDir returns exists, error
func isExistingDir(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	if err == nil && !info.IsDir() {
		err = fmt.Errorf("%q is not a directory", dir)
	}
	return err == nil, err
}
