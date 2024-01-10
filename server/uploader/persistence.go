package uploader

import (
	"context"
	"time"
)

type Persistence interface {
	CreateUploader(ctx context.Context, vals *PersistentUploader) error
	CreateUpload(ctx context.Context, upID string, vals *Upload) error

	// list of all uploaderIDs
	GetUploaderIDs(ctx context.Context) ([]string, error)

	// GetUploader with all it's uploads
	GetUploader(ctx context.Context, id string) (*PersistentUploader, error)

	// Delete the uploader and all its uploads
	DeleteUploader(ctx context.Context, id string) error

	// number of uploaders
	CountUploaders(ctx context.Context) (int, error)
}

type PersistentUploader struct {
	ID        string
	Config    Config
	CreatedAt time.Time
	Uploads   []Upload
}

func (p PersistentUploader) uploader() *Uploader {
	return &Uploader{
		id:      p.ID,
		config:  p.Config,
		created: p.CreatedAt,
	}
}
