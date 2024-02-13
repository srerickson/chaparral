package store

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/srerickson/chaparral"
	"github.com/srerickson/chaparral/server/internal/lock"
	ocfl "github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/extension"
	"github.com/srerickson/ocfl-go/ocflv1"
	"github.com/srerickson/ocfl-go/validation"
)

var (
	defaultSpec   = ocfl.Spec{1, 1}
	defaultLayout = extension.Ext0002().(extension.Layout)
)

// StorageRoot represent an existing OCFL Storage Root
type StorageRoot struct {
	id      string        // storage root's unique ID
	fs      ocfl.WriteFS  // backend
	path    string        // path for storage root in fs
	base    *ocflv1.Store // OCFL storage root
	baseErr error         // error loading OCFL storage root
	locker  *lock.Locker  // object mx
	init    *StorageInitializer
	once    sync.Once // initialize base one time
}

// StorageInitializer is used to configure new storage roots that don't exist
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
		locker: lock.NewLocker(),
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

func (store *StorageRoot) FS() ocfl.WriteFS { return store.fs }

func (store *StorageRoot) Path() string { return store.path }

func (store *StorageRoot) ID() string { return store.id }

func (store *StorageRoot) Description() string {
	if store.base != nil {
		return store.base.Description()
	}
	return ""
}

func (store *StorageRoot) Spec() ocfl.Spec {
	if store.base == nil {
		return ocfl.Spec{}
	}
	return store.base.Spec()
}

func (store *StorageRoot) ResolveID(id string) (string, error) {
	if store.base == nil {
		return "", errors.New("storage root not initialized")
	}
	return store.base.ResolveID(id)
}

// ObjectVersion represent an OCFL Object with a specific version state.
type ObjectVersion struct {
	Path    string
	ID      string
	Version int
	Head    int
	Spec    ocfl.Spec
	Alg     string
	State   chaparral.Manifest
	Message string
	User    *ocfl.User
	Created time.Time
	close   func()
}

func (objState *ObjectVersion) Close() error {
	if objState.close != nil {
		objState.close()
	}
	return nil
}

// GetObjectVersion returns an ObjectState that supports concurrent access.
// Commits to objectID will fail until the Close() is called on the returned
// Object.
func (store *StorageRoot) GetObjectVersion(ctx context.Context, objectID string, verIndex int) (*ObjectVersion, error) {
	if err := store.Ready(ctx); err != nil {
		return nil, err
	}
	unlock, err := store.locker.ReadLock(objectID)
	if err != nil {
		return nil, err
	}
	obj, err := store.base.GetObject(ctx, objectID)
	if err != nil {
		unlock()
		return nil, err
	}
	if verIndex == 0 {
		verIndex = obj.Inventory.Head.Num()
	}
	version := obj.Inventory.Version(verIndex)
	if version == nil {
		unlock()
		return nil, fmt.Errorf("version index %d not found", verIndex)
	}
	objState := ObjectVersion{
		Path:    obj.ObjectRoot.Path,
		ID:      obj.Inventory.ID,
		Alg:     obj.Inventory.DigestAlgorithm,
		Spec:    obj.Inventory.Type.Spec,
		Head:    obj.Inventory.Head.Num(),
		Version: verIndex,
		State:   map[string]chaparral.FileInfo{},
		Message: version.Message,
		User:    version.User,
		Created: version.Created,
		close:   unlock,
	}
	for d, paths := range version.State {
		objState.State[d] = chaparral.FileInfo{
			Paths:  paths,
			Fixity: obj.Inventory.GetFixity(d),
		}
	}
	return &objState, nil
}

type ObjectManifest struct {
	Path     string
	ID       string
	StoreID  string
	Alg      string
	Manifest chaparral.Manifest

	close  func()
	parent *StorageRoot
}

func (obj *ObjectManifest) OCFLManifestFixity() (ocfl.DigestMap, map[string]ocfl.DigestMap) {
	m := ocfl.DigestMap{}
	f := map[string]ocfl.DigestMap{}
	for d, info := range obj.Manifest {
		m[d] = info.Paths
		for fixAlg, fixD := range info.Fixity {
			if f[fixAlg] == nil {
				f[fixAlg] = map[string][]string{}
			}
			f[fixAlg][fixD] = append(f[fixAlg][fixD], info.Paths...)
		}

	}
	return m, f
}

func (obj *ObjectManifest) Close() error {
	if obj.close != nil {
		obj.close()
	}
	return nil
}

func (obj *ObjectManifest) GetFixity(digest string) ocfl.DigestSet {
	return obj.Manifest[digest].Fixity
}

func (obj *ObjectManifest) GetContent(digest string) (ocfl.FS, string) {
	if obj.parent == nil || obj.parent.fs == nil {
		return nil, ""
	}
	if info := obj.Manifest[digest]; len(info.Paths) > 0 {
		return obj.parent.fs, path.Join(obj.Path, info.Paths[0])
	}
	return nil, ""
}

func (store *StorageRoot) GetObjectManifest(ctx context.Context, objectID string) (*ObjectManifest, error) {
	if err := store.Ready(ctx); err != nil {
		return nil, err
	}
	unlock, err := store.locker.ReadLock(objectID)
	if err != nil {
		return nil, err
	}
	obj, err := store.base.GetObject(ctx, objectID)
	if err != nil {
		unlock()
		return nil, err
	}
	man := ObjectManifest{
		Path:     obj.ObjectRoot.Path,
		ID:       obj.Inventory.ID,
		Alg:      obj.Inventory.DigestAlgorithm,
		Manifest: chaparral.Manifest{},

		parent: store,
		close:  unlock,
	}
	for d, paths := range obj.Inventory.Manifest {
		man.Manifest[d] = chaparral.FileInfo{
			Paths:  paths,
			Fixity: obj.Inventory.GetFixity(d),
		}
	}
	return &man, nil
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
	return store.fs.RemoveAll(ctx, obj.Path)
}

func (store *StorageRoot) Validate(ctx context.Context, opts ...ocflv1.ValidationOption) (*validation.Result, error) {
	if err := store.Ready(ctx); err != nil {
		return nil, err
	}
	return store.base.Validate(ctx, opts...), nil
}

//	func (store *StorageRoot) ListObjects(ctx context.Context, objFn func(*ocflv1.Object, error) bool, concurrency int) error {
//		if err := store.Init(ctx); err != nil {
//			return err
//		}
//		setupFn := func(add func(objRoot *ocfl.ObjectRoot) bool) error {
//			return ocfl.ObjectRoots(ctx, store.group.FS, ocfl.Dir(store.path), func(obj *ocfl.ObjectRoot) error {
//				if !add(obj) {
//					return fmt.Errorf("object list interupted")
//				}
//				return nil
//			})
//		}
//		workFn := func(objRoot *ocfl.ObjectRoot) (*ocflv1.Object, error) {
//			obj := &ocflv1.Object{ObjectRoot: *objRoot}
//			return obj, obj.SyncInventory(ctx)
//		}
//		resultFn := func(objRoot *ocfl.ObjectRoot, obj *ocflv1.Object, err error) error {
//			// err from SyncInventory: if non-nil, the object has validation issues/
//			// objFn may return false to quit
//			if !objFn(obj, err) {
//				return errors.New("list objects ended prematurely")
//			}
//			return nil
//		}
//		return pipeline.Run(setupFn, workFn, resultFn, concurrency)
//	}