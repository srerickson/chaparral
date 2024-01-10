package backend

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/backend/cloud"
	"github.com/srerickson/ocfl-go/backend/local"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/s3blob"
)

type FileBackend struct {
	Path string `json:"path"`
}

func (fb *FileBackend) Name() string { return "file" }

func (fb *FileBackend) IsAccessible() (bool, error) {
	abs, err := filepath.Abs(fb.Path)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, errors.New("not a directory")
	}
	return true, nil
}

func (fb *FileBackend) NewFS() (ocfl.WriteFS, error) {
	abs, err := filepath.Abs(fb.Path)
	if err != nil {
		return nil, err
	}
	return local.NewFS(abs)
}

// supports "s3" and "azure" backend types
type S3Backend struct {
	Bucket  string     `json:"bucket"`
	Options url.Values `json:"options"`
}

func (cb S3Backend) Name() string { return "s3" }

func (cb S3Backend) IsAccessible() (bool, error) {
	ctx := context.Background()
	timeout := 1 * time.Minute // don't let this hang
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	b, err := cb.bucket()
	if err != nil {
		return false, err
	}
	return b.IsAccessible(ctx)
}

func (cb S3Backend) NewFS() (ocfl.WriteFS, error) {
	bucket, err := cb.bucket()
	if err != nil {
		return nil, err
	}
	return cloud.NewFS(bucket), nil
}

func (cb S3Backend) bucket() (*blob.Bucket, error) {
	ctx := context.Background()
	urlstr := "s3://" + cb.Bucket
	if len(cb.Options) > 0 {
		urlstr += "?" + cb.Options.Encode()

	}
	return blob.OpenBucket(ctx, urlstr)
}
