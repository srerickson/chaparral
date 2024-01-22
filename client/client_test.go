package client_test

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	chap "github.com/srerickson/chaparral/client"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/ocfl-go"
)

func TestClientNewUploader(t *testing.T) {
	testFn := func(t *testing.T, htc *http.Client, url string, store *server.StorageRoot) {
		ctx := context.Background()
		cli := chap.NewClient(htc, url)
		up, err := cli.NewUploader(ctx, []string{"sha256"}, "test")
		be.NilErr(t, err)
		defer func() {
			be.NilErr(t, cli.DeleteUploader(ctx, up.ID))
		}()
		be.Nonzero(t, up.ID)
		be.Nonzero(t, up.UploadPath)
		cont := strings.NewReader("test content")
		result, err := cli.Upload(ctx, up.UploadPath, cont)
		be.NilErr(t, err)
		be.Nonzero(t, result.Size)
	}
	testutil.RunServiceTest(t, testFn)
}

func TestClientCommit(t *testing.T) {
	testFn := func(t *testing.T, htc *http.Client, url string, store *server.StorageRoot) {
		ctx := context.Background()
		cli := chap.NewClient(htc, url)
		fixture := filepath.Join("..", "testdata", "spec-ex-full")
		obj1, obj2 := "object-01", "object-02"

		t.Run("commit v1", func(t *testing.T) {
			up, err := cli.NewUploader(ctx, []string{ocfl.SHA256}, "test v1")
			be.NilErr(t, err)
			defer func() {
				be.NilErr(t, cli.DeleteUploader(ctx, up.ID))
			}()
			stage, err := chap.NewStage(ocfl.SHA256, chap.AddDir(filepath.Join(fixture, "v1")))
			be.NilErr(t, err)
			be.NilErr(t, cli.UploadStage(ctx, up, stage))
			commit := &chap.Commit{
				StorageRootID: store.ID(),
				ObjectID:      obj1,
				State:         stage.State,
				Alg:           stage.Alg,
				Version:       1,
				User: ocfl.User{
					Name:    "A.B.",
					Address: "ab@cd.ef",
				},
				Message: "test commit 1",
			}
			be.NilErr(t, cli.CommitUploader(ctx, commit, up))
			state, err := cli.GetObjectState(ctx, store.ID(), obj1, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.StorageRootID, state.StorageRootID)
			be.Equal(t, commit.ObjectID, state.ObjectID)
			be.Equal(t, commit.Version, state.Head)
			be.Equal(t, commit.Version, state.Version)
			be.DeepEqual(t, stage.State, state.State)
			be.Equal(t, commit.Alg, state.DigestAlgorithm)
			be.Equal(t, commit.Message, state.Messsage)
			be.DeepEqual(t, commit.User, state.User)
			for _, digest := range state.State {
				f, err := cli.GetContent(ctx, store.ID(), obj1, digest, "")
				be.NilErr(t, err)
				_, err = io.Copy(io.Discard, f)
				be.NilErr(t, err)
				be.NilErr(t, f.Close())
			}
		})

		t.Run("commit v2", func(t *testing.T) {
			up, err := cli.NewUploader(ctx, []string{ocfl.SHA256}, "test v2")
			be.NilErr(t, err)
			defer func() {
				be.NilErr(t, cli.DeleteUploader(ctx, up.ID))
			}()
			stage, err := chap.NewStage(ocfl.SHA256, chap.AddDir(filepath.Join(fixture, "v2")))
			be.NilErr(t, err)
			be.NilErr(t, cli.UploadStage(ctx, up, stage))
			commit := &chap.Commit{
				StorageRootID: store.ID(),
				ObjectID:      obj1,
				State:         stage.State,
				Alg:           stage.Alg,
				Version:       2,
				User: ocfl.User{
					Name:    "C.D.",
					Address: "ef@gh.i",
				},
				Message: "test commit 2",
			}
			be.NilErr(t, cli.CommitUploader(ctx, commit, up))
			state, err := cli.GetObjectState(ctx, store.ID(), obj1, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.StorageRootID, state.StorageRootID)
			be.Equal(t, commit.ObjectID, state.ObjectID)
			be.Equal(t, commit.Version, state.Head)
			be.Equal(t, commit.Version, state.Version)
			be.DeepEqual(t, stage.State, state.State)
			be.Equal(t, commit.Alg, state.DigestAlgorithm)
			be.Equal(t, commit.Message, state.Messsage)
			be.DeepEqual(t, commit.User, state.User)
			for _, digest := range state.State {
				f, err := cli.GetContent(ctx, store.ID(), obj1, digest, "")
				be.NilErr(t, err)
				_, err = io.Copy(io.Discard, f)
				be.NilErr(t, err)
				be.NilErr(t, f.Close())
			}
		})

		t.Run("fork object", func(t *testing.T) {
			obj1State, err := cli.GetObjectState(ctx, store.ID(), obj1, 0)
			be.NilErr(t, err)
			commit := &chap.Commit{
				StorageRootID: store.ID(),
				ObjectID:      obj2,
				Version:       1,
				Alg:           obj1State.DigestAlgorithm,
				State:         obj1State.State,
				User: ocfl.User{
					Name:    "C.D.",
					Address: "ef@gh.i",
				},
				Message: "test fork",
			}
			be.NilErr(t, cli.CommitFork(ctx, commit, store.ID(), obj1))
			state, err := cli.GetObjectState(ctx, store.ID(), obj2, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.StorageRootID, state.StorageRootID)
			be.Equal(t, commit.ObjectID, state.ObjectID)
			be.Equal(t, commit.Version, state.Head)
			be.Equal(t, commit.Version, state.Version)
			be.Equal(t, commit.Alg, state.DigestAlgorithm)
			be.Equal(t, commit.Message, state.Messsage)
			be.DeepEqual(t, commit.User, state.User)
			for _, digest := range state.State {
				f, err := cli.GetContent(ctx, store.ID(), obj2, digest, "")
				be.NilErr(t, err)
				_, err = io.Copy(io.Discard, f)
				be.NilErr(t, err)
				be.NilErr(t, f.Close())
			}
			be.DeepEqual(t, obj1State.State, state.State)
		})
	}
	testutil.RunServiceTest(t, testFn)
}
