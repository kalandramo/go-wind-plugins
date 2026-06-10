package rbac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// ---------------------------------------------------------------------------
// NewEngine / Options
// ---------------------------------------------------------------------------

func TestNewEngine_Default(t *testing.T) {
	e, err := NewEngine(context.Background())
	require.Nil(t, err)
	require.NotNil(t, e)
}

func TestNewEngine_WithFullSetup(t *testing.T) {
	e, err := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithRolePermission("editor", "doc:*", "write"),
		WithUserRole("alice", "reader"),
		WithUserRole("bob", "editor"),
	)
	require.Nil(t, err)
	require.NotNil(t, e)
}

// ---------------------------------------------------------------------------
// IsAuthorized
// ---------------------------------------------------------------------------

func TestIsAuthorized_Allow(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithUserRole("alice", "reader"),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_Deny(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithUserRole("alice", "reader"),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "write", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_NoRoles(t *testing.T) {
	e, _ := NewEngine(context.Background())
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Role inheritance
// ---------------------------------------------------------------------------

func TestIsAuthorized_RoleInheritance(t *testing.T) {
	// admin inherits from editor, editor inherits from reader
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithRolePermission("editor", "doc:*", "write"),
		WithRolePermission("admin", "*", "delete"),
		// Role hierarchy: admin → editor → reader
		WithUserRole("admin", "editor"),
		WithUserRole("editor", "reader"),
		// User
		WithUserRole("alice", "admin"),
	)

	// alice should have read (via reader), write (via editor), delete (via admin)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "write", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "delete", "doc:1", "")
	assert.True(t, ok)
}

func TestIsAuthorized_NoCyclicInheritance(t *testing.T) {
	// Test cyclic role assignment doesn't cause infinite loop
	e, _ := NewEngine(context.Background(),
		WithRolePermission("roleA", "doc:1", "read"),
		WithUserRole("alice", "roleA"),
		WithUserRole("roleA", "roleB"),
		WithUserRole("roleB", "roleA"), // cycle!
	)

	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Wildcard
// ---------------------------------------------------------------------------

func TestIsAuthorized_WildcardAll(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("superadmin", "*", "*"),
		WithUserRole("root", "superadmin"),
	)
	ok, _ := e.IsAuthorized(context.Background(), "root", "anything", "everything", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Multiple roles
// ---------------------------------------------------------------------------

func TestIsAuthorized_MultipleRoles(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithRolePermission("writer", "doc:*", "write"),
		WithUserRole("alice", "reader"),
		WithUserRole("alice", "writer"),
	)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "write", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "delete", "doc:1", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// ProjectsAuthorized
// ---------------------------------------------------------------------------

func TestProjectsAuthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithUserRole("alice", "reader"),
	)
	projects := engine.MakeProjects("p1", "p2", "p3")
	result, err := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("alice"), "read", "doc:1", projects)
	require.Nil(t, err)
	assert.Len(t, result, 3)
}

func TestProjectsAuthorized_Unauthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithUserRole("alice", "reader"),
	)
	projects := engine.MakeProjects("p1", "p2")
	result, _ := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("bob"), "read", "doc:1", projects)
	assert.Len(t, result, 0)
}

// ---------------------------------------------------------------------------
// FilterAuthorizedPairs
// ---------------------------------------------------------------------------

func TestFilterAuthorizedPairs(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRolePermission("reader", "doc:*", "read"),
		WithUserRole("alice", "reader"),
	)
	pairs := engine.MakePairs(
		engine.MakePair("doc:1", "read"),
		engine.MakePair("doc:2", "write"),
		engine.MakePair("img:1", "read"),
	)
	result, _ := e.FilterAuthorizedPairs(context.Background(),
		engine.MakeSubjects("alice"), pairs)
	assert.Len(t, result, 1)
}

// ---------------------------------------------------------------------------
// SetPolicies
// ---------------------------------------------------------------------------

func TestSetPolicies(t *testing.T) {
	e, _ := NewEngine(context.Background())

	err := e.SetPolicies(context.Background(),
		engine.PolicyMap{
			"rolePermissions": map[string][]Permission{
				"admin": {{"*", "*"}},
			},
		},
		engine.RoleMap{
			"userRoles": map[string][]string{
				"root": {"admin"},
			},
		},
	)
	require.Nil(t, err)

	ok, _ := e.IsAuthorized(context.Background(), "root", "delete", "anything", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Interface compliance
// ---------------------------------------------------------------------------

func TestInterfaceCompliance(t *testing.T) {
	var _ engine.Engine = (*State)(nil)
}
