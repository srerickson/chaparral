package server

import (
	"context"
	"errors"
	"net/http"

	"strings"

	"github.com/bufbuild/connect-go"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/server/store"
	"github.com/srerickson/ocfl-go"
)

type ManageService struct {
	ManageServiceAPI
}

type ManageServiceAPI interface {
	Authorizer
	StorageRoot(id string) (*store.StorageRoot, error)
}

func (srv *ManageService) Handler() (string, http.Handler) {
	return chaparralv1connect.NewManageServiceHandler(srv)
}

func (srv *ManageService) StreamObjectRoots(ctx context.Context,
	req *connect.Request[chaparralv1.StreamObjectRootsRequest],
	stream *connect.ServerStream[chaparralv1.StreamObjectRootsResponse],
) error {
	store, err := srv.StorageRoot(req.Msg.StorageRootId)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}
	if !srv.Allowed(ctx, ActionAdminister, store.ID()) {
		err := errors.New("you don't have permssion to manage the storage root")
		return connect.NewError(connect.CodePermissionDenied, err)
	}
	each := func(obj *ocfl.ObjectRoot) error {
		// use object path relative to storage root
		var objPath = obj.Path
		if store.Path() != "." {
			objPath = strings.TrimPrefix(objPath, store.Path()+"/")
		}
		return stream.Send(&chaparralv1.StreamObjectRootsResponse{
			ObjectPath:      objPath,
			Spec:            string(obj.Spec),
			DigestAlgorithm: obj.SidecarAlg,
		})
	}
	err = ocfl.ObjectRoots(ctx, store.FS(), ocfl.Dir(store.Path()), each)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
