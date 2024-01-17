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
	base    *ocflv1.Store
	baseErr error
	group   *StorageGroup // group that the storage root is part of
	path    string        // path for the storage group within the storage group
	locker  *lock.Locker
	once    sync.Once // initialize base
}

type StorageInitializer struct {
	Create      bool      `json:"create"`
	Spec        ocfl.Spec `json:"ocfl,omitempty"`
	Description string    `json:"description,omitempty"`
	Layout      string    `json:"layout,omitempty"`
	//LayoutConfig map[string]any `json:"layout_config,omitempty"`
}

func NewStorageRoot(group *StorageGroup, path string) *StorageRoot {
	return &StorageRoot{
		group:  group,
		path:   path,
		locker: lock.NewLocker(objectMgrCap),
	}
}

func (store *StorageRoot) Ready(ctx context.Context) error {
	store.once.Do(func() {
		store.base, store.baseErr = ocflv1.GetStore(ctx, store.group.fs, store.path)
	})
	return store.baseErr
}

func (store *StorageRoot) Group() *StorageGroup { return store.group }
func (store *StorageRoot) FS() ocfl.WriteFS     { return store.group.fs }
func (store *StorageRoot) Path() string         { return store.path }

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
