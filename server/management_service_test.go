package server_test

import (
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/server"
)

var _ chaparralv1connect.ManagementServiceHandler = (*server.ManagementService)(nil)
