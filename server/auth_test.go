package server_test

import (
	"context"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
)

var _ server.Authorizer = (*server.RolePermissions)(nil)

func TestRolePermissions(t *testing.T) {
	ctx := context.Background()
	perms := testutil.DefaultRoles("main")

	be.False(t, perms.Allowed(ctx, server.ActionReadObject, "*::*"))
	be.False(t, perms.Allowed(ctx, server.ActionCommitObject, "*::*"))
	be.False(t, perms.Allowed(ctx, server.ActionDeleteObject, "*::*"))
	be.False(t, perms.Allowed(ctx, "*", "*::*"))

	// members can only read from the default storage root
	memberCtx := server.CtxWithAuthUser(ctx, testutil.MemberUser)
	be.True(t, perms.Allowed(memberCtx, server.ActionReadObject, "main::object"))
	be.True(t, perms.Allowed(memberCtx, server.ActionReadObject, "*::*"))
	be.False(t, perms.Allowed(memberCtx, server.ActionReadObject, "private::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionCommitObject, "main::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionCommitObject, "*::*"))
	be.False(t, perms.Allowed(memberCtx, server.ActionDeleteObject, "main::object"))
	be.False(t, perms.Allowed(memberCtx, server.ActionDeleteObject, "*::*"))
	be.False(t, perms.Allowed(memberCtx, "*", "*::*"))

	// manager role can do anything to objects in default storage root
	managerCtx := server.CtxWithAuthUser(ctx, testutil.ManagerUser)
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, "main::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, "*::*"))
	be.True(t, perms.Allowed(managerCtx, server.ActionReadObject, "main::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionCommitObject, "*::*"))
	be.True(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "main::object"))
	be.True(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "*::*"))
	be.True(t, perms.Allowed(managerCtx, server.ActionCommitObject, "main::object"))

	// managers can't do anything to objects outside the default storage root
	be.False(t, perms.Allowed(managerCtx, server.ActionReadObject, "private::object"))
	be.False(t, perms.Allowed(managerCtx, server.ActionCommitObject, "private::object"))
	be.False(t, perms.Allowed(managerCtx, server.ActionDeleteObject, "private::object"))
	be.False(t, perms.Allowed(managerCtx, "action", "private::object"))
	be.False(t, perms.Allowed(managerCtx, "*", "*::*"))

	// admin role can do anything
	adminCtx := server.CtxWithAuthUser(ctx, testutil.AdminUser)
	be.True(t, perms.Allowed(adminCtx, "action", "private::object"))
	be.True(t, perms.Allowed(adminCtx, "action", "main::object"))
	be.True(t, perms.Allowed(adminCtx, "*", "*::*"))
}
