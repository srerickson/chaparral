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
	tmpGroup := testutil.MkGroupTempDir(t)
	ok, err := tmpGroup.IsAccessible()
	be.NilErr(t, err)
	be.True(t, ok)
	root, err := tmpGroup.StorageRoot("test")
	be.NilErr(t, err)

	// create the storage root
	be.NilErr(t, root.Ready(ctx))

	be.True(t, root.Description() != "")
	be.Nonzero(t, root.Spec())
	_, err = root.ResolveID("test-id")
	be.NilErr(t, err)

	// the testdata storage groups has a storage root called "test" that is
	// read-only. Used here as a content source
	testdataGroup, err := testutil.MkGroupTestdata(filepath.Join("..", "testdata"))
	be.NilErr(t, err)
	ok, err = testdataGroup.IsAccessible()
	be.NilErr(t, err)
	be.True(t, ok)
	srcRoot, err := testdataGroup.StorageRoot("test")
	be.NilErr(t, err)

	srcObj, err := srcRoot.GetObject(ctx, "ark:123/abc")
	be.NilErr(t, err)
	defer srcObj.Close()
	srcState, err := srcObj.State(0)
	be.NilErr(t, err)
	id := srcObj.Inventory.ID

	// commit stage that is a fork of srcObj
	stage := ocfl.NewStage(srcObj.Inventory.DigestAlgorithm)
	stage.State = srcState.DigestMap
	stage.FS = srcObj.FS
	stage.Root = srcObj.Path
	err = stage.UnsafeSetManifestFixty(srcState.Manifest, srcObj.Inventory.Fixity)
	be.NilErr(t, err)

	// test concurrent go-routines
	errs := goGroupErrors(2, func() error {
		return root.Commit(ctx, id, stage,
			ocflv1.WithCreated(srcState.Created),
			ocflv1.WithMessage(srcState.Message),
			ocflv1.WithUser(*srcState.User),
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
	newObj, err := root.GetObject(ctx, id)
	be.NilErr(t, err)
	newState, err := newObj.State(0)
	be.NilErr(t, err)
	be.DeepEqual(t, srcState, newState)

	// delete should fail because newObj is not closed
	err = root.DeleteObject(ctx, id)
	be.True(t, errors.Is(err, lock.ErrWriteLock))

	// close newObj and check Delete()
	newObj.Close()
	errs = goGroupErrors(2, func() error {
		return root.DeleteObject(ctx, id)
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
	_, err = root.GetObject(ctx, id)
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
