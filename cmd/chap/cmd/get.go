package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
)

var get = &getCmd{
	Command: &cobra.Command{
		// TODO change this to `get --id srcpath [dstpath] and check for project
		Use:   "get [-c content-path | --inventory] [-t local-path ] object-id",
		Short: "download individual files from an object",
		Long:  ``,
	},
}

func init() {
	get.Command.Flags().StringVar(&get.objectID, "id", "", "id of object to download from (required if dir is uninitialized)")
	get.Command.Flags().BoolVar(&get.replace, "replace", false, "replace destination file if it exists")
	get.Command.Flags().IntVarP(&get.version, "version", "v", 0, "object version to download from")
	get.Command.Flags().BoolVarP(&get.pathIsContent, "content", "c", false, "treat path argument as a content path, not a logical path")

	get.Command.Run = ClientRunFunc(get)
	root.Command.AddCommand(get.Command)
}

type getCmd struct {
	*cobra.Command
	objectID      string
	version       int
	pathIsContent bool
	replace       bool
}

func (get *getCmd) Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error {
	var (
		objectID string
		storeID  string = conf.StorageRootID("")
		srcArg   string // source or content path in object
		dstArg   string
		proj     cfg.Project
		err      error
	)
	proj, err = cfg.SearchProject("")
	if err != nil && !errors.Is(err, cfg.ErrNotProject) {
		return err
	}
	switch len(args) {
	case 0:
		return errors.New("missing required argument: path of object file to download")
	case 1:
		srcArg = args[0]
		dstArg = path.Base(srcArg)
	case 2:
		srcArg = args[0]
		dstArg = args[1]
	default:
		return fmt.Errorf("unexpected argument(s) after %q", args[1])
	}
	objectID = cfg.First(get.objectID, proj.ObjectID)
	if objectID == "" {
		return errors.New("an object id must be set in a local config or with the '--id' flag")
	}
	contentPath := srcArg
	digest := ""
	if !get.pathIsContent {
		// objSrcPath is a logical path -- convert it to a content path
		state, err := cli.GetObjectVersion(ctx, storeID, objectID, get.version)
		if err != nil {
			return err
		}
		digest = state.State.PathMap()[srcArg]
		if digest == "" {
			// FIXME this shows the wrong vnum because of a bug in ocfl-go
			return fmt.Errorf("%s: file not found in %s (%s)", srcArg, objectID, get.Version)
		}
	}
	_, err = getContent(ctx, cli, storeID, objectID, digest, contentPath, dstArg, get.replace)
	return err
}

func getContent(ctx context.Context, cli *client.Client, storeID, objectID, digest, contentPath, localPath string, replace bool) (int64, error) {
	var writer io.Writer
	switch localPath {
	case "-":
		writer = os.Stdout
	default:
		perm := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if replace {
			perm = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}
		if _, err := os.Stat(localPath); err == nil {
			return 0, fmt.Errorf("%q exists", localPath)
		}
		if err := os.MkdirAll(filepath.Dir(localPath), dirMode); err != nil {
			return 0, err
		}
		fmt.Println(objectID, contentPath, "->", localPath)
		f, err := os.OpenFile(localPath, perm, fileMode)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		writer = f
	}
	reader, err := cli.GetContent(ctx, storeID, objectID, digest, contentPath)
	if err != nil {
		return 0, err
	}
	defer reader.Close()
	// TODO: better UI with download progress
	return io.Copy(writer, reader)
}
