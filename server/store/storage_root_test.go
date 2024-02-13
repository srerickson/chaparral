package store_test

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"sync"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server/internal/lock"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/ocflv1"
	"golang.org/x/exp/slices"
)

func TestStorageRoot(t *testing.T) {
	ctx := context.Background()

	// storage group backed by a tempdir
	root := testutil.NewStoreTempDir(t)
	be.NilErr(t, root.Ready(ctx))

	// create the storage root
	be.NilErr(t, root.Ready(ctx))

	be.True(t, root.Description() != "")
	be.Nonzero(t, root.Spec())
	_, err := root.ResolveID("test-id")
	be.NilErr(t, err)

	// the testdata storage groups has a storage root called "test" that is
	// read-only. Used here as a content source
	srcRoot := testutil.NewStoreTestdata(t, filepath.Join("..", "..", "testdata"))
	be.NilErr(t, srcRoot.Ready(ctx))

	srcVersion, err := srcRoot.GetObjectVersion(ctx, "ark:123/abc", 0)
	be.NilErr(t, err)
	defer srcVersion.Close()
	srcManifest, err := srcRoot.GetObjectManifest(ctx, "ark:123/abc")
	be.NilErr(t, err)
	defer srcManifest.Close()

	// commit stage that is a fork of srcObj
	stage := &ocfl.Stage{
		DigestAlgorithm: srcVersion.Alg,
	}
	stage.State = srcVersion.State.DigestMap()
	stage.ContentSource = srcManifest
	stage.FixitySource = srcManifest

	// test concurrent go-routines
	errs := goGroupErrors(2, func() error {
		return root.Commit(ctx, srcVersion.ID, stage,
			ocflv1.WithCreated(srcVersion.Created),
			ocflv1.WithMessage(srcVersion.Message),
			ocflv1.WithUser(*srcVersion.User),
			ocflv1.WithOCFLSpec(srcVersion.Spec),
		)
	})
	// should have succeeded at least once
	be.True(t, slices.Contains(errs, nil))
	// should also have failed once
	be.True(t, slices.ContainsFunc(errs, func(err error) bool {
		if err == nil {
			return false
		}
		return errors.Is(err, lock.ErrWriteLock)
	}))

	// new object exists with expected state
	newState, err := root.GetObjectVersion(ctx, srcVersion.ID, 0)
	be.NilErr(t, err)
	be.DeepEqual(t, srcVersion.State, newState.State)

	// delete should fail because newState is not closed
	err = root.DeleteObject(ctx, srcVersion.ID)
	be.True(t, errors.Is(err, lock.ErrWriteLock))

	// close newObj and check Delete()
	newState.Close()
	errs = goGroupErrors(2, func() error {
		return root.DeleteObject(ctx, newState.ID)
	})
	// DeleteObject() should have succeeded once
	be.True(t, slices.Contains(errs, nil))
	//  DeleteObject() should have failed once
	be.True(t, slices.ContainsFunc(errs, func(err error) bool {
		if err == nil {
			return false
		}
		return errors.Is(err, lock.ErrWriteLock)
	}))

	// object is gone
	_, err = root.GetObjectVersion(ctx, srcVersion.ID, 0)
	be.True(t, errors.Is(err, fs.ErrNotExist))
}

func goGroupErrors(times int, run func() error) []error {
	errs := make([]error, times)
	wg := sync.WaitGroup{}
	for i := 0; i < times; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = run()
		}()
	}
	wg.Wait()
	return errs
}
