package backend

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/backend/local"
	s3ocfl "github.com/srerickson/ocfl-go/backend/s3"
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
	Bucket  string `json:"bucket"`
	Logger  *slog.Logger
	Options url.Values `json:"options"`
}

func (cb S3Backend) Name() string { return "s3" }

func (cb S3Backend) IsAccessible() (bool, error) {
	ctx := context.Background()
	timeout := 1 * time.Minute // don't let this hang
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client, err := cb.client(ctx)
	if err != nil {
		return false, err
	}
	_, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &cb.Bucket,
		MaxKeys: aws.Int32(1),
	})
	return err == nil, err
}

func (cb S3Backend) NewFS() (ocfl.WriteFS, error) {
	ctx := context.Background()
	client, err := cb.client(ctx)
	if err != nil {
		return nil, err
	}
	return &s3ocfl.BucketFS{
		S3:     client,
		Bucket: cb.Bucket,
		Logger: cb.Logger,
	}, nil
}

func (cb S3Backend) client(ctx context.Context) (*s3.Client, error) {
	region := cb.option("region")
	endpoint := cb.option("endpoint")
	opts := []func(*config.LoadOptions) error{
		config.WithDefaultRegion(region),
	}
	if endpoint != "" {
		signingRegion := region
		if signingRegion == "" {
			signingRegion = "us-east-1"
		}
		customResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               endpoint,
					SigningRegion:     signingRegion,
					HostnameImmutable: true,
				}, nil
			})
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

func (cb S3Backend) option(key string) string {
	val := cb.Options[key]
	if len(val) > 0 {
		return val[0]
	}
	return ""
}
