package chapdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go"
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
	in := &chaparral.ObjectManifest{
		ObjectRef: chaparral.ObjectRef{
			StorageRootID: "store-id",
			ID:            "object-id",
		},
		DigestAlgorithm: "sha512",
		Spec:            "1.0",
		Path:            "a/place",
		Manifest: chaparral.Manifest{
			"abc1": chaparral.FileInfo{
				Paths: []string{"a", "b", "c"},
				Size:  13,
			},
			"abc2": chaparral.FileInfo{
				Paths:  []string{"dir/a"},
				Size:   1,
				Fixity: ocfl.DigestSet{"md5": "abc123"},
			},
		},
	}
	be.NilErr(t, chapDB.SetObjectManifest(ctx, in))
	in.Spec = "1.1"
	be.NilErr(t, chapDB.SetObjectManifest(ctx, in))
	out, err := chapDB.GetObjectManifest(ctx, in.StorageRootID, in.ID)
	be.NilErr(t, err)
	be.DeepEqual(t, in, out)
}
