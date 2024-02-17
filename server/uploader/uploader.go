package uploader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/srerickson/ocfl-go"
)

var (
	ErrUploaderNotFound = errors.New("uploader not found with the given id")
	ErrUploaderDelete   = errors.New("uploader is being deleted")
	ErrUploaderInUse    = errors.New("uploader is in use and can't be deleted")
	ErrDigestAlgorithm  = errors.New("uploader doesn't support the digest algorithm")
)

type Uploader struct {
	mgr *Manager

	// immutable
	id      string
	config  Config
	created time.Time

	// mutable
	uploads []Upload

	// sync state
	deleting bool
	refs     int
	mx       sync.RWMutex
}

func (up *Uploader) Root() (ocfl.WriteFS, string) {
	return up.mgr.fs, path.Join(up.mgr.dir, up.id)
}

func (up *Uploader) Config() *Config {
	return &Config{
		Description: up.config.Description,
		Algs:        slices.Clone(up.config.Algs),
		UserID:      up.config.UserID,
	}
}

type contentSourceFunc func(string) (ocfl.FS, string)

func (c contentSourceFunc) GetContent(digest string) (ocfl.FS, string) {
	return c(digest)
}

func (up *Uploader) ContentSource(alg string) ocfl.ContentSource {
	f := func(digest string) (ocfl.FS, string) {
		up.mx.RLock()
		defer up.mx.RUnlock()
		for _, upload := range up.uploads {
			if upload.Digests[alg] == digest {
				fs, dir := up.Root()
				return fs, path.Join(dir, upload.Name)
			}
		}
		return nil, ""
	}
	return contentSourceFunc(f)
}

type fixitySourceFunc func(string) ocfl.DigestSet

func (c fixitySourceFunc) GetFixity(digest string) ocfl.DigestSet {
	return c(digest)
}

func (up *Uploader) FixitySource(alg string) ocfl.FixitySource {
	f := func(digest string) ocfl.DigestSet {
		up.mx.RLock()
		defer up.mx.RUnlock()
		set := ocfl.DigestSet{}
		for _, upload := range up.uploads {
			if upload.Digests[alg] == digest {
				for fixalg, fixdig := range upload.Digests {
					if fixalg != alg {
						set[fixalg] = fixdig
					}
				}
			}
		}
		return set
	}
	return fixitySourceFunc(f)
}

func (up *Uploader) Created() time.Time {
	return up.created
}

func (up *Uploader) Uploads() []Upload {
	up.mx.RLock()
	defer up.mx.RUnlock()
	return up.uploads
}

func (up *Uploader) Close(ctx context.Context) error {
	// it's important to handle the mgr lock first to avoid deadlocks.
	up.mgr.mx.Lock()
	defer up.mgr.mx.Unlock()

	up.mx.Lock()
	defer up.mx.Unlock()
	if up.refs > 0 {
		up.refs--
	}
	if up.deleting {
		delete(up.mgr.uploaders, up.id)
	}
	return nil
}

func (up *Uploader) Delete(ctx context.Context) error {
	if err := up.setDelete(); err != nil {
		return err
	}
	fs, root := up.Root()
	if err := fs.RemoveAll(ctx, root); err != nil {
		return fmt.Errorf("removing files for deleted uploader %q: %w", up.id, err)
	}
	if up.mgr.persist != nil {
		if err := up.mgr.persist.DeleteUploader(ctx, up.id); err != nil {
			return fmt.Errorf("deleting persistent entries for uploader %q: %w", up.id, err)
		}
	}
	return nil
}

func (up *Uploader) setDelete() error {
	up.mx.Lock()
	defer up.mx.Unlock()
	if up.refs > 1 {
		return ErrUploaderInUse
	}
	if up.deleting {
		return ErrUploaderDelete
	}
	up.deleting = true
	return nil
}

// The Upload returned by write must not be modified!
func (up *Uploader) Write(ctx context.Context, r io.Reader) (*Upload, error) {
	up.mx.Lock()
	defer up.mx.Unlock()
	writers := make([]io.Writer, len(up.config.Algs))
	for i, alg := range up.config.Algs {
		digester := ocfl.NewDigester(alg)
		if digester == nil {
			err := fmt.Errorf("%w: %q", ErrDigestAlgorithm, alg)
			return nil, err
		}
		writers[i] = digester
	}
	name := uuid.NewString()
	fsys, uproot := up.Root()
	fullPath := path.Join(uproot, name)
	size, err := fsys.Write(ctx, fullPath, io.TeeReader(r, io.MultiWriter(writers...)))
	if err != nil {
		return nil, err
	}
	u := Upload{
		Name:    name,
		Size:    size,
		Digests: make(map[string]string, len(up.config.Algs)),
	}
	for i, alg := range up.config.Algs {
		u.Digests[alg] = writers[i].(ocfl.Digester).String()
	}
	up.uploads = append(up.uploads, u)
	if up.mgr.persist != nil {
		if err := up.mgr.persist.CreateUpload(ctx, up.id, &u); err != nil {
			return nil, fmt.Errorf("persisting upload for uploader %q: %w", up.id, err)
		}
	}

	return &u, nil
}

type Config struct {
	UserID      string   `json:"user"`
	Algs        []string `json:"digest_algorithms"`
	Description string   `json:"description"`
}

func (c Config) UsesAlg(alg string) bool {
	return slices.Contains(c.Algs, alg)
}

// Upload is an item in the Uploader manifest.
type Upload struct {
	// file name relative to uploader's root
	Name string `json:"name,omitempty"`
	// file size
	Size int64 `json:"size"`
	// digests for each algorithm used by the uploader
	Digests ocfl.DigestSet `json:"digests"`
}
