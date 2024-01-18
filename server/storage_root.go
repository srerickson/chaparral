package server

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/srerickson/chaparral/server/internal/lock"
	ocfl "github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/extension"
	"github.com/srerickson/ocfl-go/ocflv1"
	"github.com/srerickson/ocfl-go/validation"
)

const (
	objectMgrCap = 1 << 10 // max number of object references in an object manager
)

var (
	defaultSpec   = ocfl.Spec{1, 1}
	defaultLayout = extension.Ext0002().(extension.Layout)
)

// StorageRoot represent an existing OCFL Storage Root in a Group
type StorageRoot struct {
	id      string
	fs      ocfl.WriteFS
	base    *ocflv1.Store
	baseErr error
	path    string // path for storage root in backend
	locker  *lock.Locker
	init    *StorageInitializer
	once    sync.Once // initialize base
}

type StorageInitializer struct {
	Description string `json:"description,omitempty"`
	Layout      string `json:"layout,omitempty"`
	//LayoutConfig map[string]any `json:"layout_config,omitempty"`
}

func NewStorageRoot(id string, fsys ocfl.WriteFS, path string, init *StorageInitializer) *StorageRoot {
	return &StorageRoot{
		id:     id,
		fs:     fsys,
		path:   path,
		init:   init,
		locker: lock.NewLocker(objectMgrCap),
	}
}

func (store *StorageRoot) Ready(ctx context.Context) error {
	store.once.Do(func() {
		store.base, store.baseErr = ocflv1.GetStore(ctx, store.fs, store.path)
		if store.baseErr == nil || store.init == nil {
			return
		}
		store.baseErr = nil
		// try to initialize the storage root
		layout := defaultLayout
		if store.init.Layout != "" {
			l, err := extension.Get(store.init.Layout)
			if err != nil {
				store.baseErr = err
				return
			}
			layout = l.(extension.Layout)
		}
		conf := ocflv1.InitStoreConf{
			Spec:        defaultSpec,
			Description: store.init.Description,
			Layout:      layout,
		}
		if err := ocflv1.InitStore(ctx, store.fs, store.path, &conf); err != nil {
			store.baseErr = err
			return
		}
		store.base, store.baseErr = ocflv1.GetStore(ctx, store.fs, store.path)
	})
	return store.baseErr
}

func (store *StorageRoot) FS() ocfl.WriteFS {
	fs, _ := store.base.Root()
	return fs.(ocfl.WriteFS)
}

func (store *StorageRoot) Path() string { return store.path }

func (store *StorageRoot) ID() string { return store.id }

func (store *StorageRoot) Description() string {
	if store.base != nil {
		return store.base.Description()
	}
	return ""
}

func (store *StorageRoot) Spec() ocfl.Spec {
	if store.base != nil {
		return store.base.Spec()
	}
	return ocfl.Spec{}
}

func (store *StorageRoot) ResolveID(id string) (string, error) {
	if store.base == nil {
		return "", errors.New("storage root not initialized")
	}
	return store.base.ResolveID(id)
}

// ObjectState represent an OCFL Object with a specific version state.
type ObjectState struct {
	FS       ocfl.FS
	Path     string
	ID       string
	Version  int
	Head     int
	Spec     ocfl.Spec
	Alg      string
	State    ocfl.DigestMap
	Manifest ocfl.DigestMap
	Message  string
	User     *ocfl.User
	Created  time.Time
	Fixity   map[string]ocfl.DigestMap
	Close    func()
}

// GetObjectState returns an Object reference that supports concurrent access.
// Commits to objectID will fail until the Close() is called on the returned
// Object.
func (store *StorageRoot) GetObjectState(ctx context.Context, objectID string, v int) (*ObjectState, error) {
	if err := store.Ready(ctx); err != nil {
		return nil, err
	}
	unlock, err := store.locker.ReadLock(objectID)
	if err != nil {
		return nil, err
	}

	o, err := store.base.GetObject(ctx, objectID)
	if err != nil {
		unlock()
		return nil, err
	}
	if v == 0 {
		v = o.Inventory.Head.Num()
	}
	obj := ObjectState{
		FS:       o.ObjectRoot.FS,
		Path:     o.ObjectRoot.Path,
		ID:       o.Inventory.ID,
		Alg:      o.Inventory.DigestAlgorithm,
		Spec:     o.Inventory.Type.Spec,
		Head:     o.Inventory.Head.Num(),
		Manifest: o.Inventory.Manifest,
		Fixity:   o.Inventory.Fixity,
		Version:  v,
		Close:    unlock,
	}
	ver := o.Inventory.GetVersion(v)
	if ver == nil {
		unlock()
		return nil, errors.New("version not found")
	}
	obj.State = ver.State
	obj.Message = ver.Message
	obj.User = ver.User
	obj.Created = ver.Created
	// must call Close() on the returned Object.
	return &obj, nil
}

func (store *StorageRoot) Commit(ctx context.Context, objectID string, stage *ocfl.Stage, opts ...ocflv1.CommitOption) error {
	if err := store.Ready(ctx); err != nil {
		return err
	}
	unlock, err := store.locker.WriteLock(objectID)
	if err != nil {
		return err
	}
	defer unlock()
	return store.base.Commit(ctx, objectID, stage, opts...)
}

func (store *StorageRoot) DeleteObject(ctx context.Context, objectID string) error {
	if err := store.Ready(ctx); err != nil {
		return err
	}
	unlock, err := store.locker.WriteLock(objectID)
	if err != nil {
		return err
	}
	defer unlock()
	obj, err := store.base.GetObject(ctx, objectID)
	if err != nil {
		return err
	}
	writeFS, ok := obj.FS.(ocfl.WriteFS)
	if !ok {
		return errors.New("object is is read-only")
	}
	return writeFS.RemoveAll(ctx, obj.Path)
}

func (store *StorageRoot) Validate(ctx context.Context, opts ...ocflv1.ValidationOption) (*validation.Result, error) {
	if err := store.Ready(ctx); err != nil {
		return nil, err
	}
	return store.base.Validate(ctx, opts...), nil
}

// func (store *StorageRoot) ListObjects(ctx context.Context, objFn func(*ocflv1.Object, error) bool, concurrency int) error {
// 	if err := store.Init(ctx); err != nil {
// 		return err
// 	}
// 	setupFn := func(add func(objRoot *ocfl.ObjectRoot) bool) error {
// 		return ocfl.ObjectRoots(ctx, store.group.FS, ocfl.Dir(store.path), func(obj *ocfl.ObjectRoot) error {
// 			if !add(obj) {
// 				return fmt.Errorf("object list interupted")
// 			}
// 			return nil
// 		})
// 	}
// 	workFn := func(objRoot *ocfl.ObjectRoot) (*ocflv1.Object, error) {
// 		obj := &ocflv1.Object{ObjectRoot: *objRoot}
// 		return obj, obj.SyncInventory(ctx)
// 	}
// 	resultFn := func(objRoot *ocfl.ObjectRoot, obj *ocflv1.Object, err error) error {
// 		// err from SyncInventory: if non-nil, the object has validation issues/
// 		// objFn may return false to quit
// 		if !objFn(obj, err) {
// 			return errors.New("list objects ended prematurely")
// 		}
// 		return nil
// 	}
// 	return pipeline.Run(setupFn, workFn, resultFn, concurrency)
// }
