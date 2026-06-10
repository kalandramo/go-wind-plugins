package acl

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

func TestNewEngine_WithRules(t *testing.T) {
	e, err := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:1"),
		WithRule("bob", "write", "doc:2"),
	)
	require.Nil(t, err)
	require.NotNil(t, e)
}

// ---------------------------------------------------------------------------
// IsAuthorized
// ---------------------------------------------------------------------------

func TestIsAuthorized_Allow(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:1"),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_Deny(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:1"),
	)
	ok, err := e.IsAuthorized(context.Background(), "bob", "read", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_DefaultDeny(t *testing.T) {
	e, _ := NewEngine(context.Background())
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_DefaultAllow(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithDefaultAllow())
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Wildcard matching
// ---------------------------------------------------------------------------

func TestIsAuthorized_WildcardAll(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("admin", "*", "*"),
	)
	ok, err := e.IsAuthorized(context.Background(), "admin", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)

	ok, err = e.IsAuthorized(context.Background(), "admin", "write", "img:2", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_WildcardResource(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:*"),
	)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "read", "img:1", "")
	assert.False(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "write", "doc:1", "")
	assert.False(t, ok)
}

func TestIsAuthorized_WildcardAction(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "*", "doc:1"),
	)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "write", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "read", "doc:2", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Deny overrides
// ---------------------------------------------------------------------------

func TestIsAuthorized_DenyOverrides(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "*", "*"),
		WithDenyRule("alice", "delete", "doc:1"),
	)
	// allowed by wildcard but denied specifically
	ok, _ := e.IsAuthorized(context.Background(), "alice", "delete", "doc:1", "")
	assert.False(t, ok)

	// other actions still allowed
	ok, _ = e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)
}

func TestIsAuthorized_NoDenyOverride(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "*", "*"),
		WithDenyRule("alice", "delete", "doc:1"),
		WithDenyOverrides(false),
	)
	// allowed because allow is checked and deny doesn't override
	ok, _ := e.IsAuthorized(context.Background(), "alice", "delete", "doc:1", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// ProjectsAuthorized
// ---------------------------------------------------------------------------

func TestProjectsAuthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:*"),
	)
	projects := engine.MakeProjects("p1", "p2", "p3")
	result, err := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("alice"), "read", "doc:1", projects)
	require.Nil(t, err)
	// ACL doesn't distinguish projects → all match if subject is authorized
	assert.Len(t, result, 3)
}

func TestProjectsAuthorized_Unauthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:*"),
	)
	projects := engine.MakeProjects("p1", "p2")
	result, err := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("bob"), "read", "doc:1", projects)
	require.Nil(t, err)
	assert.Len(t, result, 0)
}

// ---------------------------------------------------------------------------
// FilterAuthorizedPairs
// ---------------------------------------------------------------------------

func TestFilterAuthorizedPairs(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithRule("alice", "read", "doc:*"),
	)
	pairs := engine.MakePairs(
		engine.MakePair("doc:1", "read"),
		engine.MakePair("doc:2", "write"),
		engine.MakePair("img:1", "read"),
	)
	result, err := e.FilterAuthorizedPairs(context.Background(),
		engine.MakeSubjects("alice"), pairs)
	require.Nil(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "doc:1", string(result[0].Resource))
}

// ---------------------------------------------------------------------------
// SetPolicies
// ---------------------------------------------------------------------------

func TestSetPolicies(t *testing.T) {
	e, _ := NewEngine(context.Background())

	// Initially denied
	ok, _ := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.False(t, ok)

	// Set new rules
	err := e.SetPolicies(context.Background(), engine.PolicyMap{
		"rules": []Rule{
			{Subject: "alice", Action: "read", Resource: "doc:1", Effect: "allow"},
		},
	}, nil)
	require.Nil(t, err)

	// Now allowed
	ok, _ = e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Interface compliance
// ---------------------------------------------------------------------------

func TestInterfaceCompliance(t *testing.T) {
	var _ engine.Engine = (*State)(nil)
}
