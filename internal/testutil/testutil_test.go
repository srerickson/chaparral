package testutil_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
)

func TestTestdataGroup(t *testing.T) {
	store := testutil.NewStoreTestdata(t, filepath.Join("..", "..", "testdata"))
	be.NilErr(t, store.Ready(context.Background()))
}
