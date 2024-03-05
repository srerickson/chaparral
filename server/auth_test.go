package server_test

import (
	"context"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
)

var _ server.Authorizer = (server.Roles)(nil)

func TestDefaultPermissions(t *testing.T) {
	ctx := context.Background()
	perms := server.DefaultPermissions("main")

	be.False(t, perms.Allowed(ctx, server.ActionReadObject, "*"))
	be.False(t, perms.Allowed(ctx, server.ActionCommitObject, "*"))
	be.False(t, perms.Allowed(ctx, server.ActionDeleteObject, "*"))
	be.False(t, perms.Allowed(ctx, "*", "*"))

	memberCtx := server.CtxWithAuthUser(ctx, testutil.MemberUser)
	be.True(t, perms.Allowed(memberCtx, server.ActionReadObject, "main::object"))
	be.True(t, perms.Allowed(memberCtx, server.ActionReadObject, "*"))
	be.False(t, perms.Allowed(memberCtx, server.ActionReadObject, "private::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionCommitObject, "main::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionCommitObject, "*"))
	be.False(t, perms.Allowed(memberCtx, server.ActionDeleteObject, "main::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionDeleteObject, "*"))
	be.False(t, perms.Allowed(memberCtx, "*", "*"))

	managerCtx := server.CtxWithAuthUser(ctx, testutil.ManagerUser)
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, "private::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, "*"))
	be.True(t, perms.Allowed(managerCtx, server.ActionCommitObject, "private::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionCommitObject, "*"))
	be.True(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "private::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "*"))
	be.False(t, perms.Allowed(managerCtx, "action", "private::object"))
	be.False(t, perms.Allowed(managerCtx, "*", "*"))

	adminCtx := server.CtxWithAuthUser(ctx, testutil.AdminUser)
	be.True(t, perms.Allowed(adminCtx, "action", "private::object"))
	be.True(t, perms.Allowed(adminCtx, "action", "main::object"))
	be.True(t, perms.Allowed(adminCtx, "*", "*"))
}
