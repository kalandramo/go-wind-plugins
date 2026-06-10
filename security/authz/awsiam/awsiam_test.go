package awsiam

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

func TestNewEngine_WithAllowStatement(t *testing.T) {
	e, err := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"s3:Get*"}, []string{"arn:s3:::bucket/*"}),
	)
	require.Nil(t, err)
	require.NotNil(t, e)
}

func TestNewEngine_WithPolicy(t *testing.T) {
	policy := Policy{
		Version: "2012-10-17",
		Statement: []Statement{
			{Effect: "Allow", Actions: []string{"*"}, Resources: []string{"*"}},
		},
	}
	e, err := NewEngine(context.Background(), WithPolicy("admin", policy))
	require.Nil(t, err)
	require.NotNil(t, e)
}

// ---------------------------------------------------------------------------
// IsAuthorized
// ---------------------------------------------------------------------------

func TestIsAuthorized_Allow(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"doc:read"}, []string{"doc:1"}),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "doc:read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_Deny(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"doc:read"}, []string{"doc:1"}),
	)
	ok, err := e.IsAuthorized(context.Background(), "bob", "doc:read", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_DefaultDeny(t *testing.T) {
	e, _ := NewEngine(context.Background())
	ok, _ := e.IsAuthorized(context.Background(), "alice", "doc:read", "doc:1", "")
	assert.False(t, ok)
}

func TestIsAuthorized_DefaultAllow(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithDefaultAllow())
	ok, _ := e.IsAuthorized(context.Background(), "alice", "doc:read", "doc:1", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Wildcard matching
// ---------------------------------------------------------------------------

func TestIsAuthorized_WildcardAll(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("admin", []string{"*"}, []string{"*"}),
	)
	ok, _ := e.IsAuthorized(context.Background(), "admin", "anything", "everything", "")
	assert.True(t, ok)
}

func TestIsAuthorized_WildcardPrefix(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"s3:Get*"}, []string{"arn:s3:::bucket/*"}),
	)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "s3:GetObject", "arn:s3:::bucket/file.txt", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "s3:DeleteObject", "arn:s3:::bucket/file.txt", "")
	assert.False(t, ok)
}

func TestIsAuthorized_WildcardMiddle(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"s3:*"}, []string{"arn:s3:::bucket/*/data/*"}),
	)
	ok, _ := e.IsAuthorized(context.Background(), "alice", "s3:GetObject", "arn:s3:::bucket/proj1/data/file.json", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "s3:GetObject", "arn:s3:::bucket/proj1/other/file.json", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Deny overrides
// ---------------------------------------------------------------------------

func TestIsAuthorized_DenyOverrides(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithPolicy("alice", Policy{
			Version: "2012-10-17",
			Statement: []Statement{
				{Effect: "Allow", Actions: []string{"*"}, Resources: []string{"*"}},
				{Effect: "Deny", Actions: []string{"s3:DeleteObject"}, Resources: []string{"arn:s3:::protected/*"}},
			},
		}),
	)

	// Allowed for normal action
	ok, _ := e.IsAuthorized(context.Background(), "alice", "s3:GetObject", "arn:s3:::protected/file.txt", "")
	assert.True(t, ok)

	// Denied for delete on protected
	ok, _ = e.IsAuthorized(context.Background(), "alice", "s3:DeleteObject", "arn:s3:::protected/file.txt", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Multiple statements
// ---------------------------------------------------------------------------

func TestIsAuthorized_MultipleStatements(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithPolicy("alice", Policy{
			Version: "2012-10-17",
			Statement: []Statement{
				{Effect: "Allow", Actions: []string{"doc:read"}, Resources: []string{"doc:1"}},
				{Effect: "Allow", Actions: []string{"doc:write"}, Resources: []string{"doc:2"}},
			},
		}),
	)

	ok, _ := e.IsAuthorized(context.Background(), "alice", "doc:read", "doc:1", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "doc:write", "doc:2", "")
	assert.True(t, ok)

	ok, _ = e.IsAuthorized(context.Background(), "alice", "doc:delete", "doc:1", "")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// ProjectsAuthorized
// ---------------------------------------------------------------------------

func TestProjectsAuthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"doc:read"}, []string{"doc:*"}),
	)
	result, err := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("alice"), "doc:read", "doc:1",
		engine.MakeProjects("p1", "p2", "p3"))
	require.Nil(t, err)
	assert.Len(t, result, 3)
}

func TestProjectsAuthorized_Unauthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"doc:read"}, []string{"doc:*"}),
	)
	result, _ := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("bob"), "doc:read", "doc:1",
		engine.MakeProjects("p1", "p2"))
	assert.Len(t, result, 0)
}

// ---------------------------------------------------------------------------
// FilterAuthorizedPairs
// ---------------------------------------------------------------------------

func TestFilterAuthorizedPairs(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithAllowStatement("alice", []string{"doc:read"}, []string{"doc:*"}),
	)
	pairs := engine.MakePairs(
		engine.MakePair("doc:1", "doc:read"),
		engine.MakePair("doc:2", "doc:write"),
		engine.MakePair("img:1", "doc:read"),
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
			"policies": map[string][]Policy{
				"alice": {{
					Version: "2012-10-17",
					Statement: []Statement{
						{Effect: "Allow", Actions: []string{"*"}, Resources: []string{"*"}},
					},
				}},
			},
		}, nil)
	require.Nil(t, err)

	ok, _ := e.IsAuthorized(context.Background(), "alice", "anything", "everything", "")
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// matchPattern (internal)
// ---------------------------------------------------------------------------

func TestMatchPattern_Exact(t *testing.T) {
	assert.True(t, matchPattern("doc:read", "doc:read"))
	assert.False(t, matchPattern("doc:read", "doc:write"))
}

func TestMatchPattern_Star(t *testing.T) {
	assert.True(t, matchPattern("*", "anything"))
}

func TestMatchPattern_Prefix(t *testing.T) {
	assert.True(t, matchPattern("s3:*", "s3:GetObject"))
	assert.False(t, matchPattern("s3:*", "ec2:Run"))
}

func TestMatchPattern_Middle(t *testing.T) {
	assert.True(t, matchPattern("arn:*::bucket", "arn:aws::bucket"))
	assert.False(t, matchPattern("arn:*::bucket", "arn:aws::other"))
}

// ---------------------------------------------------------------------------
// Name / Interface compliance
// ---------------------------------------------------------------------------

func TestName(t *testing.T) {
	e, _ := NewEngine(context.Background())
	assert.Equal(t, "awsiam", e.Name())
}

func TestInterfaceCompliance(t *testing.T) {
	var _ engine.Engine = (*State)(nil)
}
