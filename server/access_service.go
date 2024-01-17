package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/bufbuild/connect-go"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	AccessServiceName = chaparralv1connect.AccessServiceName
	RouteDownload     = "/" + AccessServiceName + "/" + "download"
	QueryDigest       = "digest"
	QueryContentPath  = "content_path"
	QueryObjectID     = "object_id"
)

type AccessService struct {
	*chaparral
}

func (s *AccessService) Handler() (string, http.Handler) {
	// Unlike CommiService, AccessService authorization checks
	// are handled in the hander functions.
	route, handle := chaparralv1connect.NewAccessServiceHandler(s)
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == RouteDownload && r.Method == http.MethodGet {
			s.DownloadHandler(w, r)
			return
		}
		handle.ServeHTTP(w, r)
	})
	return route, fn
}

func (s *AccessService) GetObjectState(ctx context.Context, req *connect.Request[chaparralv1.GetObjectStateRequest]) (*connect.Response[chaparralv1.GetObjectStateResponse], error) {
	logger := LoggerFromCtx(ctx).With(
		QueryStorageRoot, req.Msg.StorageRootId,
		QueryObjectID, req.Msg.ObjectId,
		"version", req.Msg.Version,
	)
	user := AuthUserFromCtx(ctx)
	store, err := s.storageRoot(req.Msg.StorageRootId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if s.auth != nil && !s.auth.RootActionAllowed(ctx, &user, ReadAction, req.Msg.GroupId, req.Msg.StorageRootId) {
		err = errors.New("you don't have permission to read from the storage root")
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	obj, err := store.GetObjectState(ctx, req.Msg.ObjectId, int(req.Msg.Version))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer obj.Close()
	resp := &chaparralv1.GetObjectStateResponse{
		GroupId:         req.Msg.GroupId,
		StorageRootId:   req.Msg.StorageRootId,
		ObjectId:        obj.ID,
		DigestAlgorithm: obj.Alg,
		Version:         int32(obj.Version),
		Head:            int32(obj.Head),
		Spec:            obj.Spec.String(),
		Messsage:        obj.Message,
		Created:         timestamppb.New(obj.Created),
		State:           obj.State.PathMap(),
	}
	if obj.User != nil {
		resp.User = (*User)(obj.User).AsProto()
	}
	return connect.NewResponse(resp), nil
}

func (srv *AccessService) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		objectRoot  string
		ctx         = r.Context()
		groupID     = r.URL.Query().Get(QueryGroupID)
		storeID     = r.URL.Query().Get(QueryStorageRoot)
		objectID    = r.URL.Query().Get(QueryObjectID)
		digest      = r.URL.Query().Get(QueryDigest)
		contentPath = r.URL.Query().Get(QueryContentPath)
		user        = AuthUserFromCtx(ctx)
		logger      = LoggerFromCtx(ctx).With(
			QueryStorageRoot, storeID,
			QueryObjectID, objectID,
			QueryDigest, digest,
		)
	)
	defer func() {
		if err != nil {
			fmt.Fprint(w, err.Error())
		}
	}()
	if contentPath == "" && digest == "" {
		err = errors.New("must provide 'content_path' or 'digest' query parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	store, err := srv.storageRoot(storeID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if srv.auth != nil && !srv.auth.RootActionAllowed(ctx, &user, ReadAction, groupID, storeID) {
		w.WriteHeader(http.StatusUnauthorized)
		err = errors.New("you don't have permission to download from the storage root")
		return
	}
	if objectID == "" {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New("malformed or missing object id")
		return
	}

	// make sure storage root's base is initialized
	if err = store.Ready(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch {
	case contentPath == "":
		// get contentPath using digest
		// TODO: use a cache
		var obj *ObjectState
		obj, err = store.GetObjectState(ctx, objectID, 0)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer obj.Close()
		objectRoot = obj.Path
		if p := obj.Manifest.DigestPaths(digest); len(p) > 0 {
			contentPath = p[0]
		}
		if contentPath == "" {
			err = fmt.Errorf("object %q has no content with digest %q", objectID, digest)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		if !fs.ValidPath(contentPath) || contentPath == "." {
			w.WriteHeader(http.StatusBadRequest)
			err = errors.New("invalid content path: " + contentPath)
			return
		}
		var objPath string
		objPath, err = store.ResolveID(objectID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		objectRoot = path.Join(store.Path(), objPath)
	}
	fullPath := path.Join(objectRoot, contentPath)
	// content path relative to group's FS
	f, err := store.FS().OpenFile(ctx, fullPath)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	if _, err = io.Copy(w, f); err != nil {
		pathErr := &fs.PathError{}
		if errors.As(err, &pathErr) && strings.HasSuffix(pathErr.Path, fullPath) {
			// only report path relative to group
			pathErr.Path = fullPath
		}
		if strings.HasSuffix(err.Error(), "is a directory") {
			// error triggered when reading a file descriptor for a directory
			// from local filesystem. This message may be os-specific.
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
