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

	be.False(t, perms.Allowed(ctx, server.ActionReadObject, ""))
	be.False(t, perms.Allowed(ctx, server.ActionCommitObject, ""))
	be.False(t, perms.Allowed(ctx, server.ActionDeleteObject, ""))
	be.False(t, perms.Allowed(ctx, "anyAction", ""))

	memberCtx := server.CtxWithAuthUser(ctx, testutil.MemberUser)
	be.True(t, perms.Allowed(memberCtx, server.ActionReadObject, ""))
	be.False(t, perms.Allowed(memberCtx, server.ActionReadObject, "anything"))
	be.False(t, perms.Allowed(memberCtx, server.ActionCommitObject, ""))
	be.False(t, perms.Allowed(memberCtx, server.ActionDeleteObject, ""))
	be.False(t, perms.Allowed(memberCtx, "anyAction", "anything"))

	managerCtx := server.CtxWithAuthUser(ctx, testutil.MemberUser)
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, ""))
	be.True(t, perms.Allowed(managerCtx, server.ActionCommitObject, ""))
	be.True(t, perms.Allowed(managerCtx, server.ActionDeleteObject, ""))
	be.False(t, perms.Allowed(managerCtx, server.ActionReadObject, "anything"))
	be.False(t, perms.Allowed(managerCtx, server.ActionCommitObject, "anything"))
	be.False(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "anyting"))
	be.False(t, perms.Allowed(managerCtx, "anyAction", "anything"))

	adminCtx := server.CtxWithAuthUser(ctx, testutil.AdminUser)
	be.True(t, perms.Allowed(adminCtx, "anyAction", "anything"))
	be.True(t, perms.Allowed(adminCtx, "anyAction", "anything"))
}
