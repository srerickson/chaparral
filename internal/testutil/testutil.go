package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/chaparral/server/backend"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go/backend/local"
	"github.com/srerickson/ocfl-go/extension"
)

var (
	s3Env     = "CHAPARRAL_TEST_S3"
	storeConf = store.StorageInitializer{
		Description: "test store",
		Layout:      extension.Ext0003().Name(),
	}
)

type ServiceTestFunc func(t *testing.T, cli *http.Client, url string, store *store.StorageRoot)

func RunServiceTest(t *testing.T, tests ...ServiceTestFunc) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	authFn := server.DefaultAuthUserFunc(&testKey().PublicKey)
	tmpData := t.TempDir()
	opts := []server.Option{
		server.WithLogger(logger),
		server.WithAuthUserFunc(authFn),
		server.WithAuthorizer(server.DefaultPermissions()),
	}
	t.Run("local-root", func(t *testing.T) {
		db, err := chapdb.Open("sqlite3", filepath.Join(tmpData, "local-db.sqlite"), true)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		defer db.Close()
		store := NewStoreTempDir(t)
		mgr := uploader.NewManager(store.FS(), "uploads", (*chapdb.SQLiteDB)(db))
		mux := server.New(append(opts,
			server.WithStorageRoots(store),
			server.WithUploaderManager(mgr))...)
		testSrv := httptest.NewTLSServer(mux)
		testCli := testSrv.Client()
		authorizeClient(testSrv.Client(), AdminUser)
		defer testSrv.Close()
		for _, ts := range tests {
			ts(t, testCli, testSrv.URL, store)
		}
	})
	if WithS3() {
		t.Run("s3-root", func(t *testing.T) {
			db, err := chapdb.Open("sqlite3", filepath.Join(tmpData, "s3-db.sqlite"), true)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			defer db.Close()
			root := NewStoreS3(t)
			mgr := uploader.NewManager(root.FS(), "uploads", (*chapdb.SQLiteDB)(db))
			mux := server.New(append(opts,
				server.WithStorageRoots(root),
				server.WithUploaderManager(mgr))...)
			testSrv := httptest.NewTLSServer(mux)
			testCli := testSrv.Client()
			authorizeClient(testSrv.Client(), AdminUser)
			defer testSrv.Close()
			for _, ts := range tests {
				ts(t, testCli, testSrv.URL, root)
			}
		})
	}
}

// Test using S3 backends
func WithS3() bool { return os.Getenv(s3Env) != "" }

func S3Session() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(os.Getenv(s3Env)),
		DisableSSL:       aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}

// Testdata storage root
func NewStoreTestdata(t *testing.T, testdataPath string) *store.StorageRoot {
	fsys, err := local.NewFS(testdataPath)
	if err != nil {
		t.Fatal(err)
	}
	dir := path.Join("storage-roots", "root-01")
	root := store.NewStorageRoot("test", fsys, dir, nil)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

// new temp directory storage root for testing
func NewStoreTempDir(t *testing.T) *store.StorageRoot {
	fsys, err := local.NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := store.NewStorageRoot("test", fsys, "ocfl", &storeConf)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

// new S3 storage root for testing
func NewStoreS3(t *testing.T) *store.StorageRoot {
	backend := S3Backend(t)
	fsys, err := backend.NewFS()
	if err != nil {
		t.Fatal(err)
	}
	root := store.NewStorageRoot("test", fsys, "ocfl", &storeConf)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

// Temp dir backend for testing
func TempDirBackend(t *testing.T) *backend.FileBackend {
	return &backend.FileBackend{Path: t.TempDir()}
}

// S3 backend with temp bucket for testing
func S3Backend(t *testing.T) *backend.S3Backend {
	bucket, err := TempBucket(t)
	if err != nil {
		t.Fatal(err)
	}
	return &backend.S3Backend{
		Bucket: bucket,
		Options: map[string][]string{
			"region":           {"us-east-1"},
			"endpoint":         {os.Getenv(s3Env)},
			"disableSSL":       {"true"},
			"s3ForcePathStyle": {"true"},
		},
	}
}

func TempBucket(t *testing.T) (string, error) {
	s3cl, err := S3Session()
	if err != nil {
		t.Fatal(err)
	}
	var bucketName string
	for {
		bucketName = randName("test-ocfld-")
		if _, err = s3cl.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		}); err == nil {
			continue
		}
		break
	}
	// create bucket
	if _, err := s3cl.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}); err != nil {
		t.Fatalf("creating bucket %q: %v", bucketName, err)
	}
	t.Cleanup(func() {
		if err := removeBucket(s3cl, bucketName); err != nil {
			slog.Debug(err.Error())
		}
	})
	return bucketName, nil
}

func removeBucket(s3cl *s3.S3, name string) error {
	b := aws.String(name)
	var listFuncErr error
	listopts := &s3.ListObjectsV2Input{Bucket: b}
	listFunc := func(out *s3.ListObjectsV2Output, last bool) bool {
		for _, obj := range out.Contents {
			if _, err := s3cl.DeleteObject(&s3.DeleteObjectInput{
				Bucket: b,
				Key:    obj.Key,
			}); err != nil {
				listFuncErr = fmt.Errorf("removing %q: %w", *obj.Key, err)
				return false
			}
		}
		return !last
	}
	if err := s3cl.ListObjectsV2Pages(listopts, listFunc); err != nil {
		return err
	}
	if listFuncErr != nil {
		return listFuncErr
	}
	_, err := s3cl.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

func randName(prefix string) string {
	byt, err := io.ReadAll(io.LimitReader(rand.Reader, 4))
	if err != nil {
		panic("randName: " + err.Error())
	}
	return prefix + hex.EncodeToString(byt)
}
