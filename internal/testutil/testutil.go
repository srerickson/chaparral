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
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go/extension"
)

var (
	s3Env     = "CHAPARRAL_TEST_S3"
	storeConf = server.StorageInitializer{
		Description: "test store",
		Layout:      extension.Ext0003().Name(),
	}
)

type ServiceTestFunc func(t *testing.T, cli *http.Client, url string, group *server.StorageRoot)

func RunServiceTest(t *testing.T, testFn ServiceTestFunc) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		//Level: slog.LevelDebug,
	}))
	testAuthFunc := server.DefaultAuthUserFunc(&testKey().PublicKey)

	tmpData := t.TempDir()
	db, err := chapdb.Open("sqlite3", filepath.Join(tmpData, "db.sqlite"), true)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	mux := server.New(
		server.WithAuthUserFunc(testAuthFunc),
		server.WithAuthorizer(server.DefaultPermissions()),
		server.WithStorageRoots(MkGroupTempDir(t)),
		// server.WithUploader()
	)
	testSrv := httptest.NewTLSServer(mux)
	testCli := testSrv.Client()
	authorizeClient(testSrv.Client(), AdminUser)
	defer testSrv.Close()
	for _, root := range roots {
		t.Run("storage-group"+root.ID(), func(t *testing.T) {
			testFn(t, testCli, testSrv.URL, root)
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

// Returns a storage group using the testdata directory as a backend.
func MkGroupTestdata(t *testing.T, testdataPath string) *server.StorageRoot {
	backend := &backend.FileBackend{Path: testdataPath}
	dir := path.Join("storage-roots", "root-01")
	root := server.NewStorageRoot("test", backend, dir, nil)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

// group config for testadata directory
func MkGroupTempDir(t *testing.T) *server.StorageRoot {
	tmpd := t.TempDir()
	backend := &backend.FileBackend{Path: tmpd}
	root := server.NewStorageRoot("test", backend, "ocfl", &storeConf)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

func FileBackend(t *testing.T) *backend.FileBackend {
	tmpd := t.TempDir()
	return &backend.FileBackend{Path: tmpd}
}

// group config for minio-based group
func MkGroupTempS3(t *testing.T) *server.StorageRoot {
	backend := S3Backend(t)
	root := server.NewStorageRoot("test", backend, "ocfl", &storeConf)
	if err := root.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	return root
}

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
