package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "github.com/tx7do/go-wind-plugins/errors"
)

// ---------------------------------------------------------------------------
// Respond — *errors.Error
// ---------------------------------------------------------------------------

func TestRespond_WrappedError(t *testing.T) {
	original := errs.New(404, "USER_NOT_FOUND", "user not found")
	wrapped := wrapErr(original)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Respond(rec, req, wrapped)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Reason != "USER_NOT_FOUND" {
		t.Errorf("reason = %q, want USER_NOT_FOUND", body.Reason)
	}
	if body.Message != "user not found" {
		t.Errorf("message = %q", body.Message)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}
}

func TestRespond_MetadataInBody(t *testing.T) {
	e := errs.New(400, "VALIDATION_ERROR", "validation failed").
		WithMetadata(map[string]string{"field": "email"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	Respond(rec, req, e)

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Meta["field"] != "email" {
		t.Errorf("meta.field = %q", body.Meta["field"])
	}
}

// ---------------------------------------------------------------------------
// Respond — plain error (not *errors.Error)
// ---------------------------------------------------------------------------

func TestRespond_PlainError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Respond(rec, req, errors.New("database connection failed"))

	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	// Should NOT leak internal error details
	if body.Message == "database connection failed" {
		t.Error("plain error details should not be leaked")
	}
	if body.Reason != "INTERNAL" {
		t.Errorf("reason = %q, want INTERNAL", body.Reason)
	}
}

// ---------------------------------------------------------------------------
// Middleware — passes through normal requests
// ---------------------------------------------------------------------------

func TestMiddleware_PassThrough(t *testing.T) {
	mw := Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want ok", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Middleware + Respond — end-to-end
// ---------------------------------------------------------------------------

func TestMiddleware_RespondEndToEnd(t *testing.T) {
	mw := Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Respond(w, r, errs.New(401, "UNAUTHORIZED", "missing token"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Code != 401 || body.Reason != "UNAUTHORIZED" {
		t.Errorf("body = %+v", body)
	}
}

// ---------------------------------------------------------------------------
// WithBodyBuilder — custom response format
// ---------------------------------------------------------------------------

type customBody struct {
	Success bool   `json:"success"`
	ErrCode string `json:"err_code"`
	Detail  string `json:"detail"`
}

func TestWithBodyBuilder(t *testing.T) {
	mw := Middleware(WithBodyBuilder(func(_ http.ResponseWriter, _ *http.Request, e *errs.Error) any {
		return customBody{
			Success: false,
			ErrCode: e.Reason,
			Detail:  e.Message,
		}
	}))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Respond(w, r, errs.New(403, "FORBIDDEN", "access denied"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	var body customBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Success != false || body.ErrCode != "FORBIDDEN" || body.Detail != "access denied" {
		t.Errorf("unexpected body: %+v", body)
	}
}

// ---------------------------------------------------------------------------
// Respond without middleware registered (fallback defaults)
// ---------------------------------------------------------------------------

func TestRespond_WithoutMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Call Respond directly without wrapping with Middleware()
	Respond(rec, req, errs.New(429, "RATE_LIMITED", "slow down"))

	if rec.Code != 429 {
		t.Fatalf("status = %d, want 429", rec.Code)
	}

	var body ErrorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Reason != "RATE_LIMITED" {
		t.Errorf("reason = %q", body.Reason)
	}
}

// ---------------------------------------------------------------------------
// Respond — nil error
// ---------------------------------------------------------------------------

func TestRespond_NilError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Respond(rec, req, nil)

	// nil error is treated as unknown → 500
	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// wrapErr wraps an error to test errors.As unwrapping.
func wrapErr(e *errs.Error) error {
	return &wrappedError{err: e}
}

type wrappedError struct {
	err error
}

func (w *wrappedError) Error() string { return "wrapped: " + w.err.Error() }
func (w *wrappedError) Unwrap() error { return w.err }
