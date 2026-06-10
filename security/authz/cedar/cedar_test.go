package cedar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// ---------------------------------------------------------------------------
// mock AVP server
// ---------------------------------------------------------------------------

func newMockAVPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		principal := req["principal"].(map[string]interface{})["entityId"].(string)
		action := req["action"].(map[string]interface{})["entityId"].(string)

		// Simple mock logic: alice can do anything, bob can only read
		allowed := false
		if principal == "alice" {
			allowed = true
		} else if principal == "bob" && action == "read" {
			allowed = true
		}

		decision := "DENY"
		if allowed {
			decision = "ALLOW"
		}

		json.NewEncoder(w).Encode(map[string]string{"decision": decision})
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

func TestNewEngine_WithPolicyStoreID(t *testing.T) {
	e, err := NewEngine(context.Background(), WithPolicyStoreID("store-123"))
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

func TestIsAuthorized_NoStoreID(t *testing.T) {
	e, _ := NewEngine(context.Background())
	_, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	assert.NotNil(t, err)
}

// ---------------------------------------------------------------------------
// IsAuthorized — remote AVP
// ---------------------------------------------------------------------------

func TestIsAuthorized_Remote_Allow(t *testing.T) {
	srv := newMockAVPServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(),
		WithPolicyStoreID("store-1"),
		WithEndpoint(srv.URL),
	)
	ok, err := e.IsAuthorized(context.Background(), "alice", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)
}

func TestIsAuthorized_Remote_Deny(t *testing.T) {
	srv := newMockAVPServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(),
		WithPolicyStoreID("store-1"),
		WithEndpoint(srv.URL),
	)
	ok, err := e.IsAuthorized(context.Background(), "charlie", "read", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_Remote_BobRead(t *testing.T) {
	srv := newMockAVPServer()
	defer srv.Close()

	e, _ := NewEngine(context.Background(),
		WithPolicyStoreID("store-1"),
		WithEndpoint(srv.URL),
	)
	ok, err := e.IsAuthorized(context.Background(), "bob", "read", "doc:1", "")
	require.Nil(t, err)
	assert.True(t, ok)

	ok, err = e.IsAuthorized(context.Background(), "bob", "write", "doc:1", "")
	require.Nil(t, err)
	assert.False(t, ok)
}

func TestIsAuthorized_Remote_ServerError(t *testing.T) {
	e, _ := NewEngine(context.Background(),
		WithPolicyStoreID("store-1"),
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
		engine.MakePair("img:1", "read"),
	)
	result, _ := e.FilterAuthorizedPairs(context.Background(),
		engine.MakeSubjects("alice"), pairs)
	assert.Len(t, result, 2)
}

// ---------------------------------------------------------------------------
// Endpoint resolution
// ---------------------------------------------------------------------------

func TestGetEndpoint_Default(t *testing.T) {
	o := &Options{}
	assert.Contains(t, o.getEndpoint(), "us-east-1")
}

func TestGetEndpoint_CustomRegion(t *testing.T) {
	o := &Options{region: "eu-west-1"}
	assert.Contains(t, o.getEndpoint(), "eu-west-1")
}

func TestGetEndpoint_CustomURL(t *testing.T) {
	o := &Options{endpoint: "http://localhost:8080"}
	assert.Equal(t, "http://localhost:8080", o.getEndpoint())
}

// ---------------------------------------------------------------------------
// Name / Interface compliance
// ---------------------------------------------------------------------------

func TestName(t *testing.T) {
	e, _ := NewEngine(context.Background())
	assert.Equal(t, "cedar", e.Name())
}

func TestInterfaceCompliance(t *testing.T) {
	var _ engine.Engine = (*State)(nil)
}

// suppress unused import
var _ = strings.HasPrefix
