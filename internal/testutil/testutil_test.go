package testutil_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
)

func TestTestdataGroup(t *testing.T) {
	group, err := testutil.MkGroupTestdata(filepath.Join("..", "..", "testdata"))
	be.NilErr(t, err)
	ok, err := group.IsAccessible()
	be.NilErr(t, err)
	be.True(t, ok)
	store, err := group.StorageRoot("test")
	be.NilErr(t, err)
	err = store.Ready(context.Background())
	be.NilErr(t, err)
}
