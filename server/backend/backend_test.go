package backend_test

import (
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/chaparral/server/backend"
)

var _ server.Backend = (*backend.S3Backend)(nil)
var _ server.Backend = (*backend.FileBackend)(nil)
