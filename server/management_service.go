package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
)

type ManagementService struct {
	*chaparral
}

func (s *ManagementService) ScanStorageRoot(ctx context.Context, req *connect.Request[chaparralv1.ScanStorageRootRequest], resp *connect.ServerStream[chaparralv1.ScanStorageRootResponse]) error {
	return errors.New("not implemented")
}
