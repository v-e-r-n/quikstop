package rbac

import (
	"context"
	"testing"
)

func TestExplicitRoles(t *testing.T) {
	rs := NewRoleSet().
		Define("viewer", "doc:read", "comment:read").
		Define("editor", "doc:read", "comment:read", "doc:write", "comment:write").
		Finalize()

	if !rs.HasPermission("viewer", "doc:read") {
		t.Errorf("expected viewer to have doc:read")
	}
	if rs.HasPermission("viewer", "doc:write") {
		t.Errorf("viewer should not have doc:write")
	}
	if !rs.HasPermission("editor", "doc:write") {
		t.Errorf("expected editor to have doc:write")
	}
}

func TestInheritedRoles(t *testing.T) {
	rs := NewRoleSet().
		Define("guest", "backlog:prioritize", "project:view").
		Define("member", "time:track").Inherits("guest").
		Define("manager", "project:manage", "invoices:create").Inherits("member").
		Define("admin", "members:invite").Inherits("manager").
		Finalize()

	// Verify manager gets member and guest permissions
	if !rs.HasPermission("manager", "backlog:prioritize") {
		t.Errorf("manager should inherit backlog:prioritize from guest")
	}
	if !rs.HasPermission("manager", "time:track") {
		t.Errorf("manager should inherit time:track from member")
	}
	if !rs.HasPermission("manager", "project:manage") {
		t.Errorf("manager should have project:manage")
	}
	if rs.HasPermission("manager", "members:invite") {
		t.Errorf("manager should NOT have admin permission members:invite")
	}

	// Verify admin inherits everything
	if !rs.HasPermission("admin", "time:track") || !rs.HasPermission("admin", "members:invite") {
		t.Errorf("admin should inherit all permissions")
	}

	perms := rs.PermissionsFor("member")
	expected := map[string]bool{
		"backlog:prioritize": true,
		"project:view":       true,
		"time:track":         true,
	}
	if len(perms) != len(expected) {
		t.Errorf("expected %d perms for member, got %d: %v", len(expected), len(perms), perms)
	}
}

func TestWildcardPermissions(t *testing.T) {
	rs := NewRoleSet().
		Define("superadmin", "*").
		Define("time_admin", "time:*").
		Finalize()

	if !rs.HasPermission("superadmin", "anything:at:all") {
		t.Errorf("superadmin with '*' should match any permission")
	}
	if !rs.HasPermission("time_admin", "time:track") || !rs.HasPermission("time_admin", "time:delete_all") {
		t.Errorf("time_admin with 'time:*' should match any time permission")
	}
	if rs.HasPermission("time_admin", "billing:manage") {
		t.Errorf("time_admin should not have billing:manage")
	}
}

func TestUndefinedRole(t *testing.T) {
	rs := NewRoleSet().Finalize()
	perms := rs.PermissionsFor("unknown")
	if perms == nil || len(perms) != 0 {
		t.Errorf("expected empty non-nil slice for undefined role, got: %v", perms)
	}
	if rs.HasPermission("unknown", "any:perm") {
		t.Errorf("undefined role should return false for HasPermission")
	}
	if err := rs.Authorize("unknown", "any:perm"); err != ErrForbidden {
		t.Errorf("expected ErrForbidden for undefined role, got: %v", err)
	}
}

func TestMultipleRoleSets(t *testing.T) {
	// Org-level roles
	orgRoles := NewRoleSet().
		Define("guest").
		Define("member", "view_dashboard").
		Define("admin", "manage_billing").Inherits("member").
		Finalize()

	// Project-level roles
	projRoles := NewRoleSet().
		Define("guest", "view_board", "prioritize_backlog").
		Define("member", "track_time", "edit_stories").Inherits("guest").
		Define("admin", "delete_project").Inherits("member").
		Finalize()

	// Guest has no org perms, but has project perms
	if orgRoles.HasPermission("guest", "view_dashboard") {
		t.Errorf("guest should not have view_dashboard in orgRoles")
	}
	if !projRoles.HasPermission("guest", "prioritize_backlog") {
		t.Errorf("guest should have prioritize_backlog in projRoles")
	}

	// Member has both
	if !orgRoles.HasPermission("member", "view_dashboard") {
		t.Errorf("member should have view_dashboard in orgRoles")
	}
	if !projRoles.HasPermission("member", "track_time") {
		t.Errorf("member should have track_time in projRoles")
	}
}

func TestContextIntegration(t *testing.T) {
	rs := NewRoleSet().
		Define("admin", "billing:manage").
		Finalize()

	ctx := context.Background()

	// Empty context
	if rs.Can(ctx, "billing:manage") {
		t.Errorf("empty context should not have permissions")
	}

	// Injected context
	ctxWithRole := WithRole(ctx, "admin")
	if !rs.Can(ctxWithRole, "billing:manage") {
		t.Errorf("expected Can to return true with injected role")
	}
}

func TestFinalizationEnforcement(t *testing.T) {
	unfinalized := NewRoleSet().Define("admin", "read", "write")

	// Evaluating before finalization must panic with ErrNotFinalized
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic when calling HasPermission on unfinalized RoleSet")
		}
		if r != ErrNotFinalized {
			t.Errorf("expected ErrNotFinalized panic, got: %v", r)
		}
	}()

	_ = unfinalized.rs.HasPermission("admin", "read")
}

func TestMutationAfterFinalizationPanics(t *testing.T) {
	rs := NewRoleSet().
		Define("admin", "read").
		Finalize()

	// Mutating after finalization must panic with ErrAlreadyFinalized
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic when mutating finalized RoleSet")
		}
		if r != ErrAlreadyFinalized {
			t.Errorf("expected ErrAlreadyFinalized panic, got: %v", r)
		}
	}()

	rs.Define("hacker", "everything")
}
