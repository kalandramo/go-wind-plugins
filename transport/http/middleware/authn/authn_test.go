package authn

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	engine "github.com/tx7do/go-wind-plugins/security/authn"
)

// fakeAuthenticator is a minimal [engine.Authenticator] for testing.
type fakeAuthenticator struct {
	token    string
	claims   *engine.AuthClaims
	failWith error
}

func (f *fakeAuthenticator) Authenticate(_ context.Context) (*engine.AuthClaims, error) {
	return f.claims, f.failWith
}

func (f *fakeAuthenticator) AuthenticateToken(token string) (*engine.AuthClaims, error) {
	if token != f.token {
		return nil, engine.ErrInvalidToken
	}
	return f.claims, nil
}

func (f *fakeAuthenticator) CreateIdentityWithContext(ctx context.Context, _ engine.AuthClaims) (context.Context, error) {
	return ctx, nil
}

func (f *fakeAuthenticator) CreateIdentity(_ engine.AuthClaims) (string, error) {
	return f.token, nil
}

func (f *fakeAuthenticator) Close() {}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func TestMiddleware_Success(t *testing.T) {
	claims := &engine.AuthClaims{engine.ClaimFieldSubject: "alice"}
	auth := &fakeAuthenticator{token: "Bearer valid-token", claims: claims}

	var capturedClaims *engine.AuthClaims
	mw := Middleware(auth)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims, _ = engine.AuthClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, capturedClaims)
	sub, _ := capturedClaims.GetSubject()
	assert.Equal(t, "alice", sub)
}

func TestMiddleware_MissingHeader(t *testing.T) {
	auth := &fakeAuthenticator{}
	mw := Middleware(auth)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddleware_AuthFailure(t *testing.T) {
	auth := &fakeAuthenticator{failWith: engine.ErrInvalidToken}
	mw := Middleware(auth)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddleware_CustomHeaderName(t *testing.T) {
	claims := &engine.AuthClaims{engine.ClaimFieldSubject: "bob"}
	auth := &fakeAuthenticator{token: "Bearer custom-token", claims: claims}

	mw := Middleware(auth, WithHeaderName("X-Auth-Token"))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// standard header → should fail
	req1 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req1.Header.Set("Authorization", "Bearer custom-token")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusUnauthorized, rec1.Code)

	// custom header → should succeed
	req2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req2.Header.Set("X-Auth-Token", "Bearer custom-token")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestMiddleware_CustomErrorHandle(t *testing.T) {
	auth := &fakeAuthenticator{failWith: engine.ErrInvalidToken}

	type errBody struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	mw := Middleware(auth, WithErrorHandle(func(w http.ResponseWriter, _ *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errBody{Code: 401, Message: err.Error()})
	}))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid bearer token")
}

func TestMiddleware_ClaimsAvailableInContext(t *testing.T) {
	claims := &engine.AuthClaims{
		engine.ClaimFieldSubject: "charlie",
		engine.ClaimFieldScope:   []string{"read", "write"},
	}
	auth := &fakeAuthenticator{token: "Bearer valid-token", claims: claims}

	var ok bool
	var ctxClaims *engine.AuthClaims

	mw := Middleware(auth)
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctxClaims, ok = engine.AuthClaimsFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.True(t, ok)
	scopes, _ := ctxClaims.GetScopes()
	assert.Equal(t, []string{"read", "write"}, []string(scopes))
}
