package cerbos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// ---------------------------------------------------------------------------
// mock Cerbos server
// ---------------------------------------------------------------------------

func newMockCerbosServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		principal := req["principal"].(map[string]interface{})["id"].(string)
		resources := req["resources"].([]interface{})
		res := resources[0].(map[string]interface{})["resource"].(map[string]interface{})
		resKind := res["kind"].(string)
		actions := resources[0].(map[string]interface{})["actions"].([]interface{})

		// Simple mock: alice can do anything on any resource
		// bob can read doc:* resources
		actionsMap := make(map[string]bool)
		for _, a := range actions {
			action := a.(string)
			allowed := false
			if principal == "alice" {
				allowed = true
			} else if principal == "bob" && action == "read" && resKind == "doc" {
				allowed = true
			}
			actionsMap[action] = allowed
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"requestId": req["requestId"],
			"results": []map[string]interface{}{
				{
					"resource": res,
					"actions":  actionsMap,
				},
			},
		})
	}))
}

// ---------------------------------------------------------------------------
// NewEngine / Options
// ---------------------------------------------------------------------------

func TestNewEngine_Default(t *testing.T) {
	e, err := NewEngine(context.Background())
	require.Nil(t, err)
	require.NotNil(t, e)
}

func TestNewEngine_WithEndpoint(t *testing.T) {
	e, err := NewEngine(context.Background(), WithEndpoint("http://localhost:3592"))
	require.Nil(t, err)
	require.NotNil(t, e)
}

// ---------------------------------------------------------------------------
// IsAuthorized — local authorizer
// ---------------------------------------------------------------------------

func TestIsAuthorized_LocalAuthorizer_Allow(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithAuthorizer(func(p, a, r string) (bool, error) {
		return p == "alice", nil
	}))
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_LocalAuthorizer_Deny(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithAuthorizer(func(p, a, r string) (bool, error) {
		return p == "alice", nil
	}))
	ok, err := e.IsAuthorized(context.Background(), "bob", "read", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// IsAuthorized — remote Cerbos
// ---------------------------------------------------------------------------

func TestIsAuthorized_Remote_Allow(t *testing.T) {
	srv := newMockCerbosServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(),
		WithEndpoint(srv.URL),
		WithPrincipalRoles(map[string][]string{"alice": {"admin"}}),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_Remote_Deny(t *testing.T) {
	srv := newMockCerbosServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(),
		WithEndpoint(srv.URL),
	)
	ok, err := e.IsAuthorized(context.Background(), "charlie", "delete", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_Remote_BobReadDoc(t *testing.T) {
	srv := newMockCerbosServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(), WithEndpoint(srv.URL))

	// bob can read doc
	ok, err := e.IsAuthorized(context.Background(), "bob", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)

	// bob cannot write
	ok, err = e.IsAuthorized(context.Background(), "bob", "write", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)

	// bob cannot read img
	ok, err = e.IsAuthorized(context.Background(), "bob", "read", "img:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_Remote_ServerError(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithEndpoint("http://127.0.0.1:0"),
	)
	_, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.NotNil(t, err)
}

// ---------------------------------------------------------------------------
// ProjectsAuthorized
// ---------------------------------------------------------------------------

func TestProjectsAuthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithAuthorizer(func(p, a, r string) (bool, error) {
		return p == "alice", nil
	}))
	result, err := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("alice"), "read", "doc:1",
		engine.MakeProjects("p1", "p2", "p3"))
	require.Nil(t, err)
	assert.Len(t, result, 3)
}

func TestProjectsAuthorized_Unauthorized(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithAuthorizer(func(p, a, r string) (bool, error) {
		return false, nil
	}))
	result, _ := e.ProjectsAuthorized(context.Background(),
		engine.MakeSubjects("alice"), "read", "doc:1",
		engine.MakeProjects("p1", "p2"))
	assert.Len(t, result, 0)
}

// ---------------------------------------------------------------------------
// FilterAuthorizedPairs
// ---------------------------------------------------------------------------

func TestFilterAuthorizedPairs(t *testing.T) {
	e, _ := NewEngine(context.Background(), WithAuthorizer(func(p, a, r string) (bool, error) {
		return a == "read", nil
	}))
	pairs := engine.MakePairs(
		engine.MakePair("doc:1", "read"),
		engine.MakePair("doc:2", "write"),
	)
	result, _ := e.FilterAuthorizedPairs(context.Background(),
		engine.MakeSubjects("alice"), pairs)
	assert.Len(t, result, 1)
}

// ---------------------------------------------------------------------------
// splitResource
// ---------------------------------------------------------------------------

func TestSplitResource(t *testing.T) {
	k, id := splitResource("doc:1")
	assert.Equal(t, "doc", k)
	assert.Equal(t, "1", id)
}

func TestSplitResource_NoColon(t *testing.T) {
	k, id := splitResource("doc")
	assert.Equal(t, "doc", k)
	assert.Equal(t, "default", id)
}

// ---------------------------------------------------------------------------
// Name / Interface compliance
// ---------------------------------------------------------------------------

func TestName(t *testing.T) {
	e, _ := NewEngine(context.Background())
	assert.Equal(t, "cerbos", e.Name())
}

func TestInterfaceCompliance(t *testing.T) {
	var _ engine.Engine = (*State)(nil)
}
