package uploader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/srerickson/ocfl-go"
)

type Manager struct {
	roots     map[string]Root
	uploaders map[string]*Uploader
	persist   Persistence
	mx        sync.Mutex
}

func NewManager(roots []Root, persist Persistence) *Manager {
	mgr := &Manager{
		roots:   make(map[string]Root, len(roots)),
		persist: persist,
	}
	for _, root := range roots {
		mgr.roots[root.ID] = root
	}
	return mgr
}

type Root struct {
	ID  string
	FS  ocfl.WriteFS
	Dir string
}

func (mgr *Manager) Root(id string) Root {
	return mgr.roots[id]
}

func (mgr *Manager) Roots() []string {
	ids := make([]string, 0, len(mgr.roots))
	for id := range mgr.roots {
		ids = append(ids, id)
	}
	return ids
}

func (mgr *Manager) NewUploader(ctx context.Context, config *Config) (string, error) {
	mgr.mx.Lock()
	defer mgr.mx.Unlock()
	if mgr.uploaders == nil {
		mgr.uploaders = map[string]*Uploader{}
	}
	id := uuid.NewString()
	upper := &Uploader{
		mgr:     mgr,
		id:      id,
		config:  *config,
		created: time.Now(),
	}
	if mgr.persist != nil {
		vals := &PersistentUploader{
			ID:        id,
			Config:    *config,
			CreatedAt: upper.created,
		}
		if err := mgr.persist.CreateUploader(ctx, vals); err != nil {
			return "", err
		}
	}
	mgr.uploaders[id] = upper
	return id, nil
}

func (mgr *Manager) GetUploader(ctx context.Context, uploaderID string) (*Uploader, error) {
	mgr.mx.Lock()
	defer mgr.mx.Unlock()
	uploader, ok := mgr.uploaders[uploaderID]
	if !ok {
		if mgr.persist == nil {
			return nil, ErrUploaderNotFound
		}
		// try to restore a previously saved uploader
		restored, err := mgr.persist.GetUploader(ctx, uploaderID)
		if err != nil {
			// TODO check error
			return nil, fmt.Errorf("%w: %v", ErrUploaderNotFound, err)
		}
		uploader = &Uploader{
			id:      restored.ID,
			config:  restored.Config,
			created: restored.CreatedAt,
			uploads: restored.Uploads,
			mgr:     mgr,
			refs:    1,
		}
		if mgr.uploaders == nil {
			mgr.uploaders = map[string]*Uploader{}
		}
		mgr.uploaders[uploader.id] = uploader
		return uploader, nil
	}
	uploader.mx.Lock()
	defer uploader.mx.Unlock()
	if uploader.deleting {
		return nil, ErrUploaderDelete
	}
	uploader.refs++
	return uploader, nil
}

// Len returns the number of uploads for a given upload root
func (mgr *Manager) Len(ctx context.Context) (int, error) {
	mgr.mx.Lock()
	defer mgr.mx.Unlock()
	if mgr.persist != nil {
		return mgr.persist.CountUploaders(ctx)
	}
	return len(mgr.uploaders), nil
}

func (mgr *Manager) UploaderIDs(ctx context.Context) ([]string, error) {
	mgr.mx.Lock()
	defer mgr.mx.Unlock()
	// read from persistence if present
	if mgr.persist != nil {
		return mgr.persist.GetUploaderIDs(ctx)
	}
	ids := make([]string, len(mgr.uploaders))
	i := 0
	for k := range mgr.uploaders {
		ids[i] = k
		i++
	}
	return ids, nil
}
