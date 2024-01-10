package cmd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"regexp"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	client "github.com/srerickson/chaparral/client"
	cfg "github.com/srerickson/chaparral/cmd/chap/config"
	delta "github.com/srerickson/chaparral/internal/diff"

	"github.com/srerickson/chaparral/cmd/chap/ui"
)

// regexp for `.chaparral` directory
var chaparralRE = regexp.MustCompile(`^(.*/)?\.` + chaparral + `$`)

var push = &pushCmd{
	Command: &cobra.Command{
		Use:   "push [--id object-id] [dir]",
		Short: "upload object version from a local directory",
		Long:  `Use push to create or update OCFL objects on the remote server using the contents a directory as the object's new state.`,
	},
}

func init() {
	push.Command.Flags().StringVar(&push.objectID, "id", "", "id of object to push to (required if dir is uninitialized)")
	push.Command.Flags().StringVarP(&push.Message, "msg", "m", "", "message to record with the new object version")
	push.Command.Flags().StringVarP(&push.User.Name, "name", "n", "", "user name to record with the new object version")
	push.Command.Flags().StringVarP(&push.User.Address, "email", "e", "", "email address to record with the new oject version")
	push.Command.Flags().IntVar(&push.Commit.Version, "version", 0, "constraint for the new object version number")
	push.Command.Flags().StringVar(&push.uploaderID, "uploader", "", "push using an existing uploader")

	push.Command.Run = ClientRunFunc(push)
	root.Command.AddCommand(push.Command)
}

type pushCmd struct {
	*cobra.Command
	objectID string
	client.Commit
	uploaderID string
}

func (cmd *pushCmd) Run(ctx context.Context, cli *client.Client, conf *cfg.Config, args []string) error {
	var (
		group    string
		store    string
		objectID string
		srcDir   string
		proj     cfg.Project
		err      error
	)
	switch {
	case len(args) > 0:
		srcDir = args[0]
		proj, err = cfg.GetProject(srcDir)
	default:
		proj, err = cfg.SearchProject("")
	}
	if err != nil && !errors.Is(err, cfg.ErrNotProject) {
		return err
	}
	err = nil
	group = conf.StorageGroupID(proj.StorageGroupID)
	store = conf.StorageRootID(proj.StorageRootID)
	objectID = cfg.First(push.objectID, proj.ObjectID)
	srcDir = cfg.First(srcDir, proj.Path())
	if srcDir == "" {
		return errors.New("could not determine the source directory with content to push")
	}
	if objectID == "" {
		return errors.New("an object id must be set in the source directory's config or with the '--id' flag")
	}
	if !proj.Empty() {
		ui.PrintValues("local object version", fmt.Sprintf("%s (%d)", proj.ObjectID, proj.Version))
	}
	cmd.Commit.GroupID = group
	cmd.Commit.StorageRootID = store
	cmd.Commit.ObjectID = objectID
	cmd.Commit.User.Address = conf.UserEmail(cmd.Commit.User.Address)
	cmd.Commit.User.Name = conf.UserName(cmd.Commit.User.Name)
	if err := doPush(ctx, cli, conf, &cmd.Commit, srcDir, cmd.uploaderID); err != nil {
		return err
	}
	// update to project's version to match commit
	// This assumes that commit's vnum was set durring doPush()
	if !proj.Empty() && proj.Version != cmd.Commit.Version {
		proj.Version = cmd.Commit.Version
		if err := proj.Save(); err != nil {
			return fmt.Errorf("updating local config: %w", err)
		}
	}
	return nil
}

func doPush(ctx context.Context, cli *client.Client, conf *cfg.Config, com *client.Commit, srcDir string, uploaderID string) error {
	// if err := cli.StorageRootExists(ctx, com.GroupID, com.StorageRootID); err != nil {
	// 	return err
	// }
	// check commit against existing object for continiuity
	existing, err := cli.GetObjectState(ctx, com.GroupID, com.StorageRootID, com.ObjectID, 0)
	if err != nil && !client.IsNotFound(err) {
		return fmt.Errorf("getting existing object state: %w", err)
	}
	switch {
	case existing == nil:
		// creating a new object
		if com.Version > 1 {
			return fmt.Errorf("object %q doesn't exist; can't create v %d", com.ObjectID, com.Version)
		}
		if com.Version == 0 {
			com.Version = 1
		}
		com.Alg = conf.DigestAlgorithm("")
	default:
		// updating an existing object
		com.Version = existing.Version + 1
		if com.Alg == "" {
			com.Alg = existing.DigestAlgorithm
		}
		if com.Alg != existing.DigestAlgorithm {
			return fmt.Errorf("existing object %q uses %s; changing to the digest algorithm to %s is not supported",
				com.ObjectID, existing.DigestAlgorithm, com.Alg)
		}
	}
	fmt.Print(ui.CommitSummary(com))
	// staging: build new object state from contents of a directory
	stg, err := ui.RunStageDir(srcDir, com.Alg, client.StageSkipDirRE(chaparralRE))
	if err != nil {
		return err
	}
	// FIXME - implement Eq for PathMap (don't use deep equal)
	if existing != nil && reflect.DeepEqual(stg.State, existing.State) {
		return errors.New("no change in object state")
	}
	com.State = stg.State

	// print diff
	var changes delta.Result
	if existing == nil {
		changes, _ = delta.Diff(nil, com.State)
	} else {
		changes, _ = delta.Diff(existing.State, com.State)
	}
	fmt.Print(ui.PrettyDiff(changes))

	// upload files (if necessary)
	// files to upload: local path -> digest
	uploadFiles := map[string]string{}
	for digest, names := range stg.Content {
		// FIXME: avoid duplicate uploads some other way.
		// if existing != nil && existing.Manifest.HasDigest(digest) {
		// 	// don't upload content that is part of existing object
		// 	continue
		// }
		uploadFiles[names[0]] = digest
	}
	var uploader *client.Uploader
	if len(uploadFiles) > 0 && uploaderID != "" {
		// try to reuse existing uploader
		uploader, err = cli.GetUploader(ctx, uploaderID)
		if err != nil {
			return err
		}
		if !slices.Contains(uploader.DigestAlgorithms, com.Alg) {
			return fmt.Errorf("uploader doesn't include %s", com.Alg)
		}
		// don't upload files that are already in the uploader
		for file, digest := range uploadFiles {
			if slices.ContainsFunc(uploader.Uploads, func(u client.Upload) bool {
				return u.Digests[com.Alg] == digest
			}) {
				delete(uploadFiles, file)
			}
		}
	}
	if len(uploadFiles) > 0 && uploader == nil {
		// try to create new uploader
		upName := path.Join(com.GroupID, com.StorageRootID, com.ObjectID)
		uploader, err = cli.NewUploader(ctx, com.GroupID, []string{com.Alg}, upName)
		if err != nil {
			return err
		}
	}
	// uploader cleanup logic
	var deleteUploader bool // whether uploader should be deleted during defer
	var pushErr error       // final push return error
	defer func() {
		if uploader == nil {
			return
		}
		if deleteUploader {
			fmt.Println("cleaning-up uploader", uploader.ID)
			var cleanupErr error
			if err := cli.DeleteUploader(ctx, uploader.ID); err != nil {
				cleanupErr = fmt.Errorf("during uploader cleanup: %w", err)
			}
			pushErr = errors.Join(pushErr, cleanupErr)
			return
		}
		fmt.Println("uploader", uploader.ID, "was not cleaned-up")
		fmt.Println("use", chaparral, "uploader --delete", uploader.ID, "to clear it manually.")
	}()

	// Run UIs: upload and commit
	pushErr = ui.RunUpload(cli, uploader, maps.Keys(uploadFiles))
	if pushErr != nil {
		if errors.Is(pushErr, ui.ErrCancled) {
			pushErr = nil
		}
		return pushErr
	}
	if uploader == nil {
		fmt.Println("no files need to be uploaded")
	}
	// confirm commit
	if !ui.RunConfirmCommit(com) {
		return pushErr
	}
	// finalize
	fmt.Println("This may take a while. You may safely close the program (with ctrl+c).")
	if err := cli.CommitUploader(ctx, com, uploader); err != nil {
		pushErr = fmt.Errorf("while writing object: %w", err)
		return pushErr
	}
	deleteUploader = true // cleanup the uploader on successful push
	return nil
}
