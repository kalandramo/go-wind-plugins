package authz

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authnEngine "github.com/tx7do/go-wind-plugins/security/authn"
	authzEngine "github.com/tx7do/go-wind-plugins/security/authz"
)

// fakeEngine is a minimal [engine.Engine] for testing.
type fakeEngine struct {
	authorized bool
	err        error
	// captured values for assertions
	capturedSubject  string
	capturedAction   string
	capturedResource string
	capturedProject  string
}

func (e *fakeEngine) Name() string { return "fake" }

func (e *fakeEngine) ProjectsAuthorized(_ context.Context, _ authzEngine.Subjects, _ authzEngine.Action, _ authzEngine.Resource, _ authzEngine.Projects) (authzEngine.Projects, error) {
	return nil, e.err
}

func (e *fakeEngine) FilterAuthorizedPairs(_ context.Context, _ authzEngine.Subjects, _ authzEngine.Pairs) (authzEngine.Pairs, error) {
	return nil, e.err
}

func (e *fakeEngine) FilterAuthorizedProjects(_ context.Context, _ authzEngine.Subjects) (authzEngine.Projects, error) {
	return nil, e.err
}

func (e *fakeEngine) IsAuthorized(_ context.Context, subject authzEngine.Subject, action authzEngine.Action, resource authzEngine.Resource, project authzEngine.Project) (bool, error) {
	e.capturedSubject = string(subject)
	e.capturedAction = string(action)
	e.capturedResource = string(resource)
	e.capturedProject = string(project)
	return e.authorized, e.err
}

func (e *fakeEngine) SetPolicies(_ context.Context, _ authzEngine.PolicyMap, _ authzEngine.RoleMap) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// requestWithClaims creates a request whose context already carries authn claims,
// simulating the authn middleware having run first.
func requestWithClaims(method, target string, claims *authnEngine.AuthClaims) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	ctx := authnEngine.ContextWithAuthClaims(r.Context(), claims)
	return r.WithContext(ctx)
}

// ---------------------------------------------------------------------------
// Middleware — basic authorize / deny
// ---------------------------------------------------------------------------

func TestMiddleware_Authorized(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	var called bool
	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "alice", eng.capturedSubject)
}

func TestMiddleware_Denied(t *testing.T) {
	eng := &fakeEngine{authorized: false}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "bob"}

	var called bool
	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMiddleware_EngineError(t *testing.T) {
	eng := &fakeEngine{err: authzEngine.ErrInvalidClaims}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "carol"}

	var called bool
	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ---------------------------------------------------------------------------
// Default resolvers
// ---------------------------------------------------------------------------

func TestMiddleware_DefaultActionResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := requestWithClaims(http.MethodPost, "/api/docs", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	// default action resolver returns the HTTP method
	assert.Equal(t, "POST", eng.capturedAction)
}

func TestMiddleware_DefaultResourceResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := requestWithClaims(http.MethodGet, "/api/docs/42", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, "/api/docs/42", eng.capturedResource)
}

// ---------------------------------------------------------------------------
// Custom resolvers
// ---------------------------------------------------------------------------

func TestMiddleware_CustomActionResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng, WithActionResolver(func(r *http.Request) string {
		switch r.Method {
		case http.MethodGet:
			return "read"
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			return "write"
		case http.MethodDelete:
			return "delete"
		}
		return ""
	}))
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, "read", eng.capturedAction)
}

func TestMiddleware_CustomResourceResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng, WithResourceResolver(func(r *http.Request) string {
		return "doc:*"
	}))
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, "doc:*", eng.capturedResource)
}

func TestMiddleware_CustomSubjectResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}

	mw := Middleware(eng, WithSubjectResolver(func(r *http.Request) string {
		return r.Header.Get("X-User")
	}))
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := httptest.NewRequest(http.MethodGet, "/api/docs/1", nil)
	r.Header.Set("X-User", "dave")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, "dave", eng.capturedSubject)
}

func TestMiddleware_CustomProjectResolver(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng, WithProjectResolver(func(r *http.Request) string {
		return r.Header.Get("X-Project")
	}))
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	r.Header.Set("X-Project", "proj-001")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, "proj-001", eng.capturedProject)
}

// ---------------------------------------------------------------------------
// Skip
// ---------------------------------------------------------------------------

func TestMiddleware_SkipFunc(t *testing.T) {
	eng := &fakeEngine{authorized: false}

	var called bool
	mw := Middleware(eng, WithSkipFunc(func(r *http.Request) bool {
		return r.URL.Path == "/healthz"
	}))
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// /healthz should be skipped
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)

	// /api/data should NOT be skipped
	called = false
	r2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, r2)
	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rec2.Code)
}

// ---------------------------------------------------------------------------
// Custom error handler
// ---------------------------------------------------------------------------

func TestMiddleware_CustomErrorHandle(t *testing.T) {
	eng := &fakeEngine{authorized: false}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	type errBody struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	mw := Middleware(eng, WithErrorHandle(func(w http.ResponseWriter, _ *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(errBody{Code: 403, Message: err.Error()})
	}))
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called")
	}))

	r := requestWithClaims(http.MethodGet, "/api/docs/1", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "context missing authz claims")
}

// ---------------------------------------------------------------------------
// No authn claims in context
// ---------------------------------------------------------------------------

func TestMiddleware_NoAuthnClaims(t *testing.T) {
	eng := &fakeEngine{authorized: true}

	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	// No authn claims in context — subject will be empty.
	r := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	// The engine is still called, subject is empty string.
	assert.Equal(t, "", eng.capturedSubject)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------------
// Authz claims injected into context
// ---------------------------------------------------------------------------

func TestMiddleware_AuthzClaimsInContext(t *testing.T) {
	eng := &fakeEngine{authorized: true}
	claims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "alice"}

	mw := Middleware(eng)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		authzClaims, ok := authzEngine.AuthClaimsFromContext(r.Context())
		require.True(t, ok)
		require.NotNil(t, authzClaims)
		require.NotNil(t, authzClaims.Subject)
		assert.Equal(t, "alice", string(*authzClaims.Subject))
		require.NotNil(t, authzClaims.Action)
		assert.Equal(t, "GET", string(*authzClaims.Action))
		require.NotNil(t, authzClaims.Resource)
		assert.Equal(t, "/api/data", string(*authzClaims.Resource))
	}))

	r := requestWithClaims(http.MethodGet, "/api/data", claims)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------------
// Chained: authn → authz
// ---------------------------------------------------------------------------

func TestMiddleware_ChainedWithAuthn(t *testing.T) {
	// Simulate a full authn + authz chain using fake components.
	authnClaims := &authnEngine.AuthClaims{authnEngine.ClaimFieldSubject: "eve"}
	eng := &fakeEngine{authorized: true}

	var finalSubject string

	// Fake authn handler: injects claims, then calls next.
	authnHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authnEngine.ContextWithAuthClaims(r.Context(), authnClaims)
		// In real life, authn.Middleware does this. Here we simulate.
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// authz middleware
			authzMw := Middleware(eng)
			authzMw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				// final handler
				authnC, _ := authnEngine.AuthClaimsFromContext(r.Context())
				finalSubject, _ = authnC.GetSubject()
			})).ServeHTTP(w, r)
		}).ServeHTTP(w, r.WithContext(ctx))
	})

	r := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	authnHandler.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "eve", finalSubject)
	assert.Equal(t, "eve", eng.capturedSubject)
}
