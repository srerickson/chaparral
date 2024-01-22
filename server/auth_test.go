package server_test

import (
	"context"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
)

var _ server.Authorizer = (server.Permissions)(nil)

func TestDefaultPermissions(t *testing.T) {
	ctx := context.Background()
	perms := server.DefaultPermissions()

	be.False(t, perms.RootActionAllowed(ctx, &server.AuthUser{}, server.ReadAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &server.AuthUser{}, server.CommitAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &server.AuthUser{}, server.DeleteAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &server.AuthUser{}, server.AdminAction, ""))

	be.True(t, perms.RootActionAllowed(ctx, &testutil.MemberUser, server.ReadAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.MemberUser, server.ReadAction, "anything"))

	be.False(t, perms.RootActionAllowed(ctx, &testutil.MemberUser, server.CommitAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.MemberUser, server.DeleteAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.MemberUser, server.AdminAction, ""))

	be.True(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.ReadAction, ""))
	be.True(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.CommitAction, ""))
	be.True(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.DeleteAction, ""))

	be.False(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.AdminAction, ""))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.ReadAction, "anything"))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.CommitAction, "anything"))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.DeleteAction, "anyting"))
	be.False(t, perms.RootActionAllowed(ctx, &testutil.ManagerUser, server.AdminAction, "anything"))

	be.True(t, perms.ActionAllowed(ctx, &testutil.AdminUser, "anyting"))
	be.True(t, perms.RootActionAllowed(ctx, &testutil.AdminUser, "anything", "anyting"))
}
