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
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/ocflv1"
	"golang.org/x/exp/slices"
)

func TestStorageRoot(t *testing.T) {
	roots := []*store.StorageRoot{
		testutil.NewStoreTempDir(t),
	}
	if testutil.WithS3() {
		roots = append(roots, testutil.NewStoreS3(t))
	}
	for _, r := range roots {
		testObjectLifecycle(t, r)
	}
}

func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	srcID := "ark:123/abc"
	root := testutil.NewStoreTestdata(t, filepath.Join("..", "..", "testdata"))
	if _, err := root.GetObjectManifest(ctx, srcID); err != nil {
		t.Fatal(err)
	}
	times := 20
	wg := sync.WaitGroup{}
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, err := root.GetObjectManifest(ctx, srcID)
			if err != nil {
				t.Error(err)
			}
			for _, info := range m.Manifest {
				if len(info.Paths) < 1 {
					t.Error("manifest has empty path entries")
				}
			}
		}()
	}
	wg.Wait()
}

func testObjectLifecycle(t *testing.T, root *store.StorageRoot) {
	ctx := context.Background()

	// create the storage root
	be.NilErr(t, root.Ready(ctx))

	be.True(t, root.Description() != "")
	be.Nonzero(t, root.Spec())
	_, err := root.ResolveID("test-id")
	be.NilErr(t, err)

	// the testdata storage group has a storage root called "test" that is
	// read-only. Used here as a content source
	srcID := "ark:123/abc"
	srcRoot := testutil.NewStoreTestdata(t, filepath.Join("..", "..", "testdata"))
	be.NilErr(t, srcRoot.Ready(ctx))

	// commit state from version
	srcVersion, err := srcRoot.GetObjectVersion(ctx, srcID, 0)
	be.NilErr(t, err)
	defer srcVersion.Close()

	// need manifest for content/fixity source
	srcManifest, err := srcRoot.GetObjectManifest(ctx, srcID)
	be.NilErr(t, err)
	defer srcManifest.Close()

	// commit stage that is a fork of srcObj
	stage := &ocfl.Stage{
		DigestAlgorithm: srcVersion.DigestAlgorithm,
		State:           srcVersion.State.DigestMap(),
		ContentSource:   srcManifest,
		FixitySource:    srcManifest,
	}

	// test concurrent go-routines
	errs := goGroupErrors(2, func() error {
		return root.Commit(ctx, srcID, stage,
			ocflv1.WithCreated(srcVersion.Created),
			ocflv1.WithMessage(srcVersion.Message),
			ocflv1.WithUser(*srcVersion.User),
			// ocflv1.WithOCFLSpec(srcVersion.Spec),
		)
	})
	// should have succeeded only once
	be.True(t, slices.Contains(errs, nil))
	be.True(t, slices.ContainsFunc(errs, func(err error) bool {
		return err != nil
	}))

	// new object exists with expected state
	newVer, err := root.GetObjectVersion(ctx, srcID, 0)
	be.NilErr(t, err)
	be.DeepEqual(t, srcVersion.State, newVer.State)

	// delete should fail because newState is not closed
	err = root.DeleteObject(ctx, srcID)
	be.True(t, errors.Is(err, lock.ErrWriteLock))

	// close newObj and check Delete()
	newVer.Close()
	errs = goGroupErrors(2, func() error {
		return root.DeleteObject(ctx, newVer.ObjectID)
	})
	// DeleteObject() should have succeeded only once
	be.True(t, slices.Contains(errs, nil))
	be.True(t, slices.ContainsFunc(errs, func(err error) bool {
		return err != nil
	}))
	// object is gone
	_, err = root.GetObjectVersion(ctx, srcVersion.ObjectID, 0)
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
