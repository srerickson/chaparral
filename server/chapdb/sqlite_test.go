package chapdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/uploader"
)

func TestSQLiteDB(t *testing.T) {
	ctx := context.Background()
	db, err := chapdb.Open("sqlite3", ":memory:", true)
	be.NilErr(t, err)
	defer db.Close()
	chapBD := (*chapdb.SQLiteDB)(db)

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
		err = chapBD.CreateUploader(ctx, input)
		be.NilErr(t, err)
		output, err := chapBD.GetUploader(ctx, "uploader")
		be.NilErr(t, err)
		be.DeepEqual(t, input, output)

		ids, err := chapBD.GetUploaderIDs(ctx)
		be.NilErr(t, err)
		be.DeepEqual(t, []string{"uploader"}, ids)

		be.NilErr(t, chapBD.DeleteUploader(ctx, "uploader"))
		counter, err := chapBD.CountUploaders(ctx)
		be.NilErr(t, err)
		be.Equal(t, 0, counter)
	})
}
