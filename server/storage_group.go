package server

import ()

// type StorageGroup struct {
// 	id      string
// 	backend Backend
// 	fs      ocfl.WriteFS
// 	stores  map[string]*StorageRoot

// 	allowUploads bool
// 	uploadRoot   string
// }

// func NewStorageGroup(id string, backend Backend) (*StorageGroup, error) {
// 	fsys, err := backend.NewFS()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &StorageGroup{
// 		fs:      fsys,
// 		id:      id,
// 		backend: backend,
// 		stores:  make(map[string]*StorageRoot),
// 	}, nil
// }

// func (g *StorageGroup) AddStorageRoot(id string, root string) error {
// 	if g.stores[id] != nil {
// 		return fmt.Errorf("storage root %q already exists", id)
// 	}
// 	if !fs.ValidPath(root) {
// 		return fmt.Errorf("invalid storage root path %q", root)
// 	}
// 	existingRoots := make([]string, 0, len(g.stores))
// 	for _, store := range g.stores {
// 		existingRoots = append(existingRoots, store.path)
// 	}
// 	if g.allowUploads {
// 		existingRoots = append(existingRoots, g.uploadRoot)
// 	}
// 	if pathConflict(append(existingRoots, root)...) {
// 		return fmt.Errorf("%q conflicts with storage group paths: %s", root, strings.Join(existingRoots, ", "))
// 	}
// 	g.stores[id] = NewStorageRoot(g, root)
// 	return nil
// }

// func (g *StorageGroup) SetUploadRoot(uploadRootDir string) error {
// 	if !fs.ValidPath(uploadRootDir) {
// 		return errors.New("invalid path for upload root: " + uploadRootDir)
// 	}
// 	existingRoots := make([]string, 0, len(g.stores))
// 	for _, store := range g.stores {
// 		existingRoots = append(existingRoots, store.path)
// 	}
// 	if pathConflict(append(existingRoots, uploadRootDir)...) {
// 		return fmt.Errorf("%q conflicts with storage group paths: %q", uploadRootDir, strings.Join(existingRoots, ", "))
// 	}
// 	g.uploadRoot = uploadRootDir
// 	g.allowUploads = true
// 	return nil
// }

// func (g *StorageGroup) UploadRoot() *uploader.Root {
// 	if !g.allowUploads {
// 		return nil
// 	}
// 	return &uploader.Root{
// 		ID:  g.id,
// 		FS:  g.fs,
// 		Dir: g.uploadRoot,
// 	}
// }

// // InitStorageRoot creates a new storage root and adds it to the storage group
// func (g *StorageGroup) InitStorageRoot(ctx context.Context, id string, root string, conf *ocflv1.InitStoreConf) error {
// 	if err := g.AddStorageRoot(id, root); err != nil {
// 		return err
// 	}
// 	if _, err := ocflv1.GetStore(ctx, g.fs, root); err == nil {
// 		return nil
// 	}
// 	if conf == nil {
// 		conf = &ocflv1.InitStoreConf{
// 			Spec:        defaultSpec,
// 			Description: "",
// 			Layout:      defaultLayout,
// 		}
// 	}
// 	if err := ocflv1.InitStore(ctx, g.fs, root, conf); err != nil {
// 		err = fmt.Errorf("initializing a new storage root %q in %q: %w", root, g.ID(), err)
// 		return err
// 	}
// 	return nil
// }

// // ID returns the group's name
// func (g *StorageGroup) ID() string {
// 	return g.id
// }

// func (g *StorageGroup) IsAccessible() (bool, error) {
// 	return g.backend.IsAccessible()
// }

// func (group *StorageGroup) StorageRoots() []string {
// 	stores := maps.Keys(group.stores)
// 	sort.Strings(stores)
// 	return stores
// }

// func (g *StorageGroup) StorageRoot(name string) (*StorageRoot, error) {
// 	store := g.stores[name]
// 	if store == nil {
// 		return nil, ErrStorageRootNotFound
// 	}
// 	return store, nil
// }
