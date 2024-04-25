package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/bufbuild/connect-go"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/server/store"
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

func (srv *ManageService) IndexObject(ctx context.Context,
	req *connect.Request[chaparralv1.IndexObjectRequest],
) (*connect.Response[chaparralv1.IndexObjectResponse], error) {
	if !srv.Allowed(ctx, ActionAdminister, req.Msg.StorageRootId) {
		err := errors.New("you don't have permssion to manage the storage root")
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	_, err := srv.StorageRoot(req.Msg.StorageRootId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	// store.GetObjectManifest(y)
	return nil, errors.New("not implemented")
}

func (srv *ManageService) IndexStorageRoot(ctx context.Context,
	req *connect.Request[chaparralv1.IndexStorageRootRequest],
) (*connect.Response[chaparralv1.IndexStorageRootResponse], error) {
	if !srv.Allowed(ctx, ActionAdminister, req.Msg.StorageRootId) {
		err := errors.New("you don't have permssion to manage the storage root")
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	store, err := srv.StorageRoot(req.Msg.StorageRootId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	go store.Reindex(context.WithoutCancel(ctx))
	return &connect.Response[chaparralv1.IndexStorageRootResponse]{}, nil
}
