package server_test

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
	srcRoot := testutil.NewStoreTestdata(t, filepath.Join("..", "testdata"))
	be.NilErr(t, srcRoot.Ready(ctx))

	srcObj, err := srcRoot.GetObjectState(ctx, "ark:123/abc", 0)
	be.NilErr(t, err)
	defer srcObj.Close()

	// commit stage that is a fork of srcObj
	stage := ocfl.NewStage(srcObj.Alg)
	stage.State = srcObj.State
	stage.FS = srcObj.FS
	stage.Root = srcObj.Path
	err = stage.UnsafeSetManifestFixty(srcObj.Manifest, srcObj.Fixity)
	be.NilErr(t, err)

	// test concurrent go-routines
	errs := goGroupErrors(2, func() error {
		return root.Commit(ctx, srcObj.ID, stage,
			ocflv1.WithCreated(srcObj.Created),
			ocflv1.WithMessage(srcObj.Message),
			ocflv1.WithUser(*srcObj.User),
			ocflv1.WithOCFLSpec(srcObj.Spec),
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
	newState, err := root.GetObjectState(ctx, srcObj.ID, 0)
	be.NilErr(t, err)
	be.DeepEqual(t, srcObj.State, newState.State)

	// delete should fail because newState is not closed
	err = root.DeleteObject(ctx, srcObj.ID)
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
	_, err = root.GetObjectState(ctx, srcObj.ID, 0)
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
