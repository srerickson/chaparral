package chapdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/uploader"
)

func TestUploader(t *testing.T) {
	ctx := context.Background()
	db, err := chapdb.Open("sqlite3", ":memory:", true)
	be.NilErr(t, err)
	defer db.Close()
	chapDB := (*chapdb.SQLiteDB)(db)

	t.Run("round-trip", func(t *testing.T) {
		// check that GetUploader returns what was given to CreateUploader
		input := &uploader.PersistentUploader{
			ID:        "uploader",
			CreatedAt: time.Now().UTC(),
			Config: uploader.Config{
				UserID:      "user",
				Algs:        []string{"sha512", "md5"},
				Description: "description",
			},
			Uploads: []uploader.Upload{
				{Name: "file", Size: 12, Digests: map[string]string{"a": "b"}},
			},
		}
		err = chapDB.CreateUploader(ctx, input)
		be.NilErr(t, err)
		output, err := chapDB.GetUploader(ctx, "uploader")
		be.NilErr(t, err)
		be.DeepEqual(t, input, output)

		ids, err := chapDB.GetUploaderIDs(ctx)
		be.NilErr(t, err)
		be.DeepEqual(t, []string{"uploader"}, ids)

		be.NilErr(t, chapDB.DeleteUploader(ctx, "uploader"))
		counter, err := chapDB.CountUploaders(ctx)
		be.NilErr(t, err)
		be.Equal(t, 0, counter)
	})
}

func TestObject(t *testing.T) {
	ctx := context.Background()
	db, err := chapdb.Open("sqlite3", ":memory:", true)
	be.NilErr(t, err)
	defer db.Close()
	chapDB := (*chapdb.SQLiteDB)(db)
	be.NilErr(t, chapDB.SetObject(ctx, "store", "obj", "./path", 1, "ocfl_v0.1", "sha512"))
	be.NilErr(t, chapDB.SetObject(ctx, "store", "obj", "./path2", 2, "ocfl_v1.1", "sha256"))
	p, err := chapDB.GetObject(ctx, "store", "obj")
	be.NilErr(t, err)
	be.Equal(t, "./path2", p)
}
