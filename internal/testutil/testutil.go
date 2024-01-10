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
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/chaparral/server/backend"
	"github.com/srerickson/chaparral/server/chapdb"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/extension"
	"github.com/srerickson/ocfl-go/ocflv1"
)

var s3Env = "CHAPARRAL_TEST_S3"

type ServiceTestFunc func(t *testing.T, cli *http.Client, url string, group *server.StorageGroup)

func RunServiceTest(t *testing.T, testFn ServiceTestFunc) {
	grps := []*server.StorageGroup{MkGroupTempDir(t)}
	if WithS3() {
		grps = append(grps, MkGroupTempS3(t))
	}
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
		server.WithStorageGroups(grps...),
		server.WithSQLDB(db),
	)
	testSrv := httptest.NewTLSServer(mux)
	testCli := testSrv.Client()
	authorizeClient(testSrv.Client(), AdminUser)
	defer testSrv.Close()
	for _, grp := range grps {
		t.Run("storage-group"+grp.ID(), func(t *testing.T) {
			testFn(t, testCli, testSrv.URL, grp)
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
func MkGroupTestdata(testdataPath string) (*server.StorageGroup, error) {
	grp, err := server.NewStorageGroup("testdata", &backend.FileBackend{Path: testdataPath})
	if err != nil {
		return nil, err
	}
	if err := grp.AddStorageRoot("test", path.Join("storage-roots", "root-01")); err != nil {
		return nil, err
	}
	return grp, nil
}

// group config for testadata directory
func MkGroupTempDir(t *testing.T) *server.StorageGroup {
	tmpd := t.TempDir()
	back := &backend.FileBackend{Path: tmpd}
	grpID := strings.ReplaceAll(tmpd, "/", "-")
	grp, err := server.NewStorageGroup(grpID, back)
	if err != nil {
		t.Fatal(err)
	}
	storeConf := &ocflv1.InitStoreConf{
		Spec:        ocfl.Spec1_1,
		Description: "test store",
		Layout:      extension.Ext0003().(extension.Layout),
	}
	if err := grp.InitStorageRoot(context.Background(), "test", "test", storeConf); err != nil {
		t.Fatal(err)
	}
	if err := grp.SetUploadRoot("uploads"); err != nil {
		t.Fatal(err)
	}
	return grp
}

func FileBackend(t *testing.T) *backend.FileBackend {
	tmpd := t.TempDir()
	return &backend.FileBackend{Path: tmpd}
}

// group config for minio-based group
func MkGroupTempS3(t *testing.T) *server.StorageGroup {
	back := S3Backend(t)
	grp, err := server.NewStorageGroup("s3-"+back.Bucket, back)
	if err != nil {
		t.Fatal(err)
	}
	storeConf := &ocflv1.InitStoreConf{
		Spec:        ocfl.Spec1_1,
		Description: "test store w/ S3 backend",
		Layout:      extension.Ext0003().(extension.Layout),
	}
	if err := grp.InitStorageRoot(context.Background(), "test", "test", storeConf); err != nil {
		t.Fatal(err)
	}
	if err := grp.SetUploadRoot("uploads"); err != nil {
		t.Fatal(err)
	}
	return grp
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
