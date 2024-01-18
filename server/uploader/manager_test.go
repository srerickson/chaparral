package uploader_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go"
)

const (
	desc = "test uploader"
	user = "user-10"
)

var algs = []string{ocfl.SHA256, ocfl.MD5}

func TestManager(t *testing.T) {
	ctx := context.Background()

	// persist uploaders to sqlite file
	tmpDir := t.TempDir()
	db, err := chapdb.Open("sqlite3", filepath.Join(tmpDir, "db.sqlite"), true)
	be.NilErr(t, err)
	persist := (*chapdb.SQLiteDB)(db)
	fileBack := testutil.FileBackend(t)
	fsys, err := fileBack.NewFS()
	be.NilErr(t, err)

	if testutil.WithS3() {
		s3Back := testutil.S3Backend(t)
		fsys, err = s3Back.NewFS()
		be.NilErr(t, err)
		roots = append(roots, []uploader.Root{
			{ID: "s3-root1", FS: fsys, Dir: "root1"},
			{ID: "s3-root2", FS: fsys, Dir: "root2"},
		}...)
	}
	mgr := uploader.NewManager(roots, persist)
	for _, id := range mgr.Roots() {
		uploadRootID := id
		t.Run(uploadRootID, func(t *testing.T) {
			t.Parallel()
			uploaderID, err := mgr.NewUploader(ctx, &uploader.Config{
				UserID:      user,
				Algs:        algs,
				Description: desc,
			})
			be.NilErr(t, err)
			size, err := mgr.Len(ctx)
			be.NilErr(t, err)
			be.Nonzero(t, size)
			testManagerUploaders(t, mgr, uploaderID)
		})
	}
}

func testManagerUploaders(t *testing.T, mgr *uploader.Manager, uploaderID string) {
	ctx := context.Background()
	numUploads := 10
	for i := 0; i < numUploads; i++ {
		i := i
		testName := fmt.Sprintf("upload-%d", i)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			upper, err := mgr.GetUploader(ctx, uploaderID)
			defer func() {
				be.NilErr(t, upper.Close(ctx))
			}()
			be.NilErr(t, err)
			result, err := upper.Write(ctx, strings.NewReader(testName))
			be.NilErr(t, err)
			be.True(t, len(result.Digests) == len(algs))
			be.Nonzero(t, result.Name)
			be.Nonzero(t, result.Size)
			uploads := upper.Uploads()
			be.True(t, slices.ContainsFunc(uploads, func(u uploader.Upload) bool {
				return u.Name == result.Name
			}))
			beExixtingFiles(t, upper, result.Name)
		})
	}
	t.Run("delete", func(t *testing.T) {
		// wait until uploads complete and test delete
		t.Parallel()
		upper, err := mgr.GetUploader(ctx, uploaderID)
		be.NilErr(t, err)
		timeout := 5 * time.Second
		start := time.Now()
		for {
			if time.Since(start) > timeout {
				upper.Close(ctx)
				t.Fatal("test timed-out without deleting uploader")
			}
			uploads := upper.Uploads()
			if len(uploads) < numUploads {
				// wait for uploads to finish
				time.Sleep(10 * time.Microsecond)
				continue
			}
			// try to delete...
			if err := upper.Delete(ctx); err == nil {
				// close the uploader
				be.NilErr(t, upper.Close(ctx))
				beDeletedFiles(t, upper)
				// uploaderID should exist in manager
				if _, err := mgr.GetUploader(ctx, uploaderID); err == nil {
					t.Fatal("uploader wasn't deleted")
				}
				// files should be gone
				break
			}
		}
	})

}

func beDeletedFiles(t *testing.T, upper *uploader.Uploader) {
	ctx := context.Background()
	fsys, dir := upper.Root()
	entries, err := fsys.ReadDir(ctx, dir)
	if err != nil && (!errors.Is(err, fs.ErrNotExist)) {
		t.Fatalf("unexpected error, got %v", err)
	}
	if n := len(entries); n > 0 {
		t.Fatalf("%s has %d entries", dir, n)
	}
}

func beExixtingFiles(t *testing.T, upper *uploader.Uploader, name string) {
	ctx := context.Background()
	fsys, dir := upper.Root()
	f, err := fsys.OpenFile(ctx, path.Join(dir, name))
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
}
