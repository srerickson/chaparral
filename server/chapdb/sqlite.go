package chapdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/srerickson/chaparral"
	sqlite "github.com/srerickson/chaparral/server/chapdb/sqlite_gen"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go"
)

var (
	sqliteOpts = url.Values{
		"_journal": {"WAL"},
		"_sync":    {"NORMAL"},
		"_timeout": {"5000"},
	}
)

type SQLiteDB sql.DB

func Open(driver string, file string, migrate bool) (*sql.DB, error) {
	var db *sql.DB
	switch driver {
	case "sqlite3":
		var err error
		connStr := file + "?" + sqliteOpts.Encode()
		slog.Debug("sqlite3 db", "file", connStr)
		db, err = sql.Open(driver, file+"?"+url.Values(sqliteOpts).Encode())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported driver %q", driver)
	}

	// create tables
	if migrate {
		if err := Migrate(db); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func (db *SQLiteDB) sqlDB() *sql.DB {
	return (*sql.DB)(db)
}

func (db *SQLiteDB) CreateUploader(ctx context.Context, upper *uploader.PersistentUploader) error {
	qry := sqlite.New(db.sqlDB())
	_, err := qry.CreateUploader(ctx, sqlite.CreateUploaderParams{
		ID:          upper.ID,
		UserID:      upper.Config.UserID,
		Algs:        strings.Join(upper.Config.Algs, ","),
		Description: upper.Config.Description,
		CreatedAt:   upper.CreatedAt.UTC(),
	})
	if err != nil {
		return err
	}
	for _, up := range upper.Uploads {
		if err := db.CreateUpload(ctx, upper.ID, &up); err != nil {
			return err
		}
	}
	return nil
}

func (db *SQLiteDB) CreateUpload(ctx context.Context, upID string, up *uploader.Upload) error {
	qry := sqlite.New(db.sqlDB())
	digBytes, err := json.Marshal(up.Digests)
	if err != nil {
		return err
	}
	_, err = qry.CreateUpload(ctx, sqlite.CreateUploadParams{
		ID:         up.Name,
		UploaderID: upID,
		Size:       up.Size,
		Digests:    digBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

// list of all uploaderIDs
func (db *SQLiteDB) GetUploaderIDs(ctx context.Context) ([]string, error) {
	qry := sqlite.New(db.sqlDB())
	return qry.GetUploaderIDs(ctx)
}

// GetUploader with all it's uploads
func (db *SQLiteDB) GetUploader(ctx context.Context, id string) (*uploader.PersistentUploader, error) {
	qry := sqlite.New(db.sqlDB())
	sqlUpper, err := qry.GetUploader(ctx, id)
	if err != nil {
		return nil, err
	}
	upper := &uploader.PersistentUploader{
		ID:        sqlUpper.ID,
		CreatedAt: sqlUpper.CreatedAt.UTC(),
		Config: uploader.Config{
			UserID:      sqlUpper.UserID,
			Algs:        strings.Split(sqlUpper.Algs, ","),
			Description: sqlUpper.Description,
		},
	}
	sqlUps, err := qry.GetUploads(ctx, id)
	if err != nil {
		return nil, err
	}
	upper.Uploads = make([]uploader.Upload, len(sqlUps))
	for i, u := range sqlUps {
		var digests ocfl.DigestSet
		if err := json.Unmarshal(u.Digests, &digests); err != nil {
			return nil, err
		}
		upper.Uploads[i] = uploader.Upload{
			Name:    u.ID,
			Size:    u.Size,
			Digests: digests,
		}
	}
	return upper, nil
}

// Delete the uploader and all its uploads
func (db *SQLiteDB) DeleteUploader(ctx context.Context, id string) error {
	qry := sqlite.New(db.sqlDB())

	if err := qry.DeleteUploads(ctx, id); err != nil {
		return err
	}

	if err := qry.DeleteUploader(ctx, id); err != nil {
		return err
	}
	return nil
}

// number of uploaders
func (db *SQLiteDB) CountUploaders(ctx context.Context) (int, error) {
	qry := sqlite.New(db.sqlDB())
	n, err := qry.CountUploaders(ctx)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (db *SQLiteDB) SetObjectManifest(ctx context.Context, obj *chaparral.ObjectManifest) error {
	qry := sqlite.New(db.sqlDB())
	dbObj, err := qry.CreateObject(ctx, sqlite.CreateObjectParams{
		StoreID: obj.StorageRootID,
		OcflID:  obj.ObjectID,
		Path:    obj.Path,
		Spec:    obj.Spec,
		Alg:     obj.DigestAlgorithm,
	})
	if err != nil {
		return err
	}
	for digest, info := obj.Manifest {
		_, err = qry.CreateObjectContent(ctx, sqlite.CreateObjectContentParams{
			ObjectID: dbObj.ID,
			Digest: digest,

		})
	}

	
	return nil
}

func (db *SQLiteDB) GetObject(ctx context.Context, storeID, objID string) (string, error) {
	qry := sqlite.New(db.sqlDB())
	obj, err := qry.GetObject(ctx, sqlite.GetObjectParams{
		StoreID: storeID,
		OcflID:  objID,
	})
	if err != nil {
		return "", err
	}
	return obj.Path, nil
}
