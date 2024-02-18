package chaparral_test

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	chap "github.com/srerickson/chaparral"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/ocfl-go"
)

func TestClientNewUploader(t *testing.T) {
	testFn := func(t *testing.T, htc *http.Client, url string, store *store.StorageRoot) {
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
	testFn := func(t *testing.T, htc *http.Client, url string, store *store.StorageRoot) {
		ctx := context.Background()
		cli := chap.NewClient(htc, url)
		fixture := filepath.Join("testdata", "spec-ex-full")
		obj1ID, obj2ID := "object-01", "object-02"

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
				To: chap.ObjectRef{
					StorageRootID: store.ID(),
					ID:            obj1ID,
				},
				State:   stage.State,
				Alg:     stage.Alg,
				Version: 1,
				User: ocfl.User{
					Name:    "A.B.",
					Address: "ab@cd.ef",
				},
				Message:        "test commit 1",
				ContentSources: []any{up.UploaderRef},
			}
			be.NilErr(t, cli.Commit(ctx, commit))
			state, err := cli.GetObjectVersion(ctx, store.ID(), obj1ID, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.To.StorageRootID, state.StorageRootID)
			be.Equal(t, commit.To.ID, state.ID)
			be.Equal(t, commit.Version, state.Head)
			be.Equal(t, commit.Version, state.Version)
			be.DeepEqual(t, stage.State, state.State.PathMap())
			be.Equal(t, commit.Alg, state.DigestAlgorithm)
			be.Equal(t, commit.Message, state.Message)
			if state.User != nil {
				be.DeepEqual(t, commit.User, *state.User)
			}
			for digest := range state.State {
				f, err := cli.GetContent(ctx, store.ID(), obj1ID, digest, "")
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
				To: chap.ObjectRef{
					StorageRootID: store.ID(),
					ID:            obj1ID,
				},
				State:   stage.State,
				Alg:     stage.Alg,
				Version: 2,
				User: ocfl.User{
					Name:    "C.D.",
					Address: "ef@gh.i",
				},
				Message:        "test commit 2",
				ContentSources: []any{up.UploaderRef},
			}
			be.NilErr(t, cli.Commit(ctx, commit))
			state, err := cli.GetObjectVersion(ctx, store.ID(), obj1ID, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.To.StorageRootID, state.StorageRootID)
			be.Equal(t, commit.To.ID, state.ID)
			be.Equal(t, commit.Version, state.Head)
			be.Equal(t, commit.Version, state.Version)
			be.DeepEqual(t, stage.State, state.State.PathMap())
			be.Equal(t, commit.Alg, state.DigestAlgorithm)
			be.Equal(t, commit.Message, state.Message)
			if state.User != nil {
				be.DeepEqual(t, commit.User, *state.User)
			}
			for digest := range state.State {
				f, err := cli.GetContent(ctx, store.ID(), obj1ID, digest, "")
				be.NilErr(t, err)
				_, err = io.Copy(io.Discard, f)
				be.NilErr(t, err)
				be.NilErr(t, f.Close())
			}
		})

		t.Run("fork object", func(t *testing.T) {
			// created obj2 as fork of obj1's last version
			// expected
			obj1, err := cli.GetObjectVersion(ctx, store.ID(), obj1ID, 0)
			be.NilErr(t, err)
			commit := &chap.Commit{
				To: chap.ObjectRef{
					StorageRootID: store.ID(),
					ID:            obj2ID,
				},
				Version: 1,
				Alg:     obj1.DigestAlgorithm,
				State:   obj1.State.PathMap(),
				User: ocfl.User{
					Name:    "C.D.",
					Address: "ef@gh.i",
				},
				Message: "test fork",
				ContentSources: []any{
					chap.ObjectRef{StorageRootID: store.ID(), ID: obj1ID},
				},
			}
			be.NilErr(t, cli.Commit(ctx, commit))
			// result
			obj2, err := cli.GetObjectVersion(ctx, store.ID(), obj2ID, 0)
			be.NilErr(t, err)
			be.Equal(t, commit.To.StorageRootID, obj2.StorageRootID)
			be.Equal(t, commit.To.ID, obj2.ID)
			be.Equal(t, commit.Version, obj2.Head)
			be.Equal(t, commit.Version, obj2.Version)
			be.Equal(t, commit.Alg, obj2.DigestAlgorithm)
			be.Equal(t, commit.Message, obj2.Message)
			if obj2.User != nil {
				be.DeepEqual(t, commit.User, *obj2.User)
			}
			for digest := range obj2.State {
				f, err := cli.GetContent(ctx, store.ID(), obj2ID, digest, "")
				be.NilErr(t, err)
				_, err = io.Copy(io.Discard, f)
				be.NilErr(t, err)
				be.NilErr(t, f.Close())
			}
			be.DeepEqual(t, obj1.State, obj2.State)
		})
	}
	testutil.RunServiceTest(t, testFn)
}
