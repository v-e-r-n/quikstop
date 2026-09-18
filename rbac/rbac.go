package rbac

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var (
	// ErrForbidden is returned when a role lacks the requested permission.
	ErrForbidden = errors.New("forbidden")
	// ErrNotFinalized is returned or panicked when a RoleSet is evaluated before being finalized.
	ErrNotFinalized = errors.New("roleset is not finalized")
	// ErrAlreadyFinalized is returned or panicked when attempting to define roles on an already finalized RoleSet.
	ErrAlreadyFinalized = errors.New("roleset is already finalized")
)

// RoleSet manages role definitions, permissions, and inheritance.
// A RoleSet must be Finalized before any evaluation queries (HasPermission, PermissionsFor, Can, Authorize).
type RoleSet struct {
	mu        sync.RWMutex
	roles     map[string]*Role
	finalized bool
}

// Role represents a defined role with explicit permissions and inherited parent roles.
type Role struct {
	Name        string
	Permissions map[string]struct{}
	Parents     []string
}

// NewRoleSet creates an empty mutable RoleSet.
func NewRoleSet() *RoleSet {
	return &RoleSet{
		roles: make(map[string]*Role),
	}
}

// RoleBuilder provides a fluent interface for configuring a role within a RoleSet.
type RoleBuilder struct {
	rs   *RoleSet
	role *Role
}

// Define registers or updates a role with an explicit list of permissions.
// Panics if the RoleSet is already finalized.
func (rs *RoleSet) Define(roleName string, permissions ...string) *RoleBuilder {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.finalized {
		panic(ErrAlreadyFinalized)
	}

	r, exists := rs.roles[roleName]
	if !exists {
		r = &Role{
			Name:        roleName,
			Permissions: make(map[string]struct{}),
			Parents:     nil,
		}
		rs.roles[roleName] = r
	}

	for _, p := range permissions {
		if p != "" {
			r.Permissions[p] = struct{}{}
		}
	}

	return &RoleBuilder{rs: rs, role: r}
}

// Inherits configures the role to inherit permissions from one or more parent roles.
// Panics if the RoleSet is already finalized.
func (b *RoleBuilder) Inherits(parentRoles ...string) *RoleBuilder {
	b.rs.mu.Lock()
	defer b.rs.mu.Unlock()

	if b.rs.finalized {
		panic(ErrAlreadyFinalized)
	}

	for _, parent := range parentRoles {
		if parent != "" {
			b.role.Parents = append(b.role.Parents, parent)
		}
	}
	return b
}

// Define continues the fluent chain to define another role.
func (b *RoleBuilder) Define(roleName string, permissions ...string) *RoleBuilder {
	return b.rs.Define(roleName, permissions...)
}

// Finalize freezes the RoleSet for evaluation and prevents any subsequent mutations.
// Returns the finalized *RoleSet ready for evaluation.
func (rs *RoleSet) Finalize() *RoleSet {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.finalized = true
	return rs
}

// Finalize freezes the underlying RoleSet from the builder chain.
func (b *RoleBuilder) Finalize() *RoleSet {
	return b.rs.Finalize()
}

// IsFinalized returns true if the RoleSet has been finalized.
func (rs *RoleSet) IsFinalized() bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.finalized
}

// PermissionsFor returns all effective permissions for a role, resolving any inheritance.
// Panics with ErrNotFinalized if the RoleSet has not been finalized.
// Returns an empty non-nil slice if the role is not defined.
func (rs *RoleSet) PermissionsFor(roleName string) []string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if !rs.finalized {
		panic(ErrNotFinalized)
	}

	permMap := make(map[string]struct{})
	visited := make(map[string]bool)

	var collect func(name string)
	collect = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		r, exists := rs.roles[name]
		if !exists {
			return
		}

		for p := range r.Permissions {
			permMap[p] = struct{}{}
		}
		for _, parent := range r.Parents {
			collect(parent)
		}
	}

	collect(roleName)

	perms := make([]string, 0, len(permMap))
	for p := range permMap {
		perms = append(perms, p)
	}
	return perms
}

// HasPermission checks if a role directly or indirectly grants the given permission.
// Panics with ErrNotFinalized if the RoleSet has not been finalized.
func (rs *RoleSet) HasPermission(roleName, permission string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if !rs.finalized {
		panic(ErrNotFinalized)
	}

	visited := make(map[string]bool)

	var check func(name string) bool
	check = func(name string) bool {
		if visited[name] {
			return false
		}
		visited[name] = true

		r, exists := rs.roles[name]
		if !exists {
			return false
		}

		if _, ok := r.Permissions[permission]; ok {
			return true
		}
		// Check wildcard permissions (e.g. "time:*" matches "time:track", "*" matches all)
		if _, ok := r.Permissions["*"]; ok {
			return true
		}
		parts := strings.Split(permission, ":")
		if len(parts) > 1 && len(parts[0]) > 0 {
			if _, ok := r.Permissions[parts[0]+":*"]; ok {
				return true
			}
		}

		for _, parent := range r.Parents {
			if check(parent) {
				return true
			}
		}
		return false
	}

	return check(roleName)
}

// Authorize returns nil if the role grants the permission, or ErrForbidden.
// Returns ErrNotFinalized if the RoleSet has not been finalized.
func (rs *RoleSet) Authorize(roleName, permission string) error {
	rs.mu.RLock()
	if !rs.finalized {
		rs.mu.RUnlock()
		return ErrNotFinalized
	}
	rs.mu.RUnlock()

	if rs.HasPermission(roleName, permission) {
		return nil
	}
	return ErrForbidden
}

// -----------------------------------------------------------------------------
// Context Integration
// -----------------------------------------------------------------------------

type roleContextKey struct{}

// WithRole injects a role into context.Context.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleContextKey{}, role)
}

// RoleFromContext extracts the role from context.Context.
func RoleFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(roleContextKey{}).(string); ok {
		return val
	}
	return ""
}

// Can checks if the role stored in context satisfies the permission according to rs.
func (rs *RoleSet) Can(ctx context.Context, permission string) bool {
	role := RoleFromContext(ctx)
	if role == "" {
		return false
	}
	return rs.HasPermission(role, permission)
}
