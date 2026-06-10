package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	e := New(404, "USER_NOT_FOUND", "user not found")
	want := "USER_NOT_FOUND: user not found"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_Error_NoReason(t *testing.T) {
	e := New(500, "", "something broke")
	want := "something broke"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestNewf(t *testing.T) {
	e := Newf(400, "INVALID_FIELD", "field %q is required", "email")
	if e.Message != `field "email" is required` {
		t.Errorf("Message = %q", e.Message)
	}
	if e.Code != 400 || e.Reason != "INVALID_FIELD" {
		t.Errorf("Code=%d, Reason=%q", e.Code, e.Reason)
	}
}

func TestFromError_Direct(t *testing.T) {
	original := New(404, "USER_NOT_FOUND", "user not found")
	got := FromError(original)
	if got != original {
		t.Error("FromError should return the same *Error")
	}
}

func TestFromError_Wrapped(t *testing.T) {
	original := New(404, "USER_NOT_FOUND", "user not found")
	wrapped := fmt.Errorf("service layer: %w", original)
	got := FromError(wrapped)
	if got == nil {
		t.Fatal("expected non-nil *Error from wrapped error")
	}
	if got.Code != 404 || got.Reason != "USER_NOT_FOUND" {
		t.Errorf("got Code=%d Reason=%q", got.Code, got.Reason)
	}
}

func TestFromError_DoubleWrapped(t *testing.T) {
	original := New(400, "BAD_REQUEST", "bad input")
	wrapped1 := fmt.Errorf("handler: %w", original)
	wrapped2 := fmt.Errorf("middleware: %w", wrapped1)
	got := FromError(wrapped2)
	if got == nil || got.Code != 400 || got.Reason != "BAD_REQUEST" {
		t.Errorf("expected nested unwrap to succeed, got %+v", got)
	}
}

func TestFromError_NotAnError(t *testing.T) {
	got := FromError(stderrors.New("plain error"))
	if got != nil {
		t.Error("expected nil for non-*Error")
	}
}

func TestFromError_Nil(t *testing.T) {
	got := FromError(nil)
	if got != nil {
		t.Error("expected nil for nil error")
	}
}

func TestCode_FromError(t *testing.T) {
	e := New(403, "FORBIDDEN", "access denied")
	if code := Code(e); code != 403 {
		t.Errorf("Code() = %d, want 403", code)
	}
}

func TestCode_FromWrappedError(t *testing.T) {
	e := New(429, "RATE_LIMITED", "too many requests")
	wrapped := fmt.Errorf("gateway: %w", e)
	if code := Code(wrapped); code != 429 {
		t.Errorf("Code() = %d, want 429", code)
	}
}

func TestCode_PlainError(t *testing.T) {
	if code := Code(stderrors.New("boom")); code != 500 {
		t.Errorf("Code() = %d, want 500", code)
	}
}

func TestCode_NilError(t *testing.T) {
	if code := Code(nil); code != 200 {
		t.Errorf("Code() = %d, want 200", code)
	}
}

func TestWithMetadata(t *testing.T) {
	original := New(400, "VALIDATION_ERROR", "validation failed")
	clone := original.WithMetadata(map[string]string{
		"field": "email",
	})

	// Clone should have metadata
	if clone.Metadata["field"] != "email" {
		t.Errorf("expected field=email, got %q", clone.Metadata["field"])
	}

	// Original should be unchanged
	if original.Metadata != nil {
		t.Error("original should not have metadata")
	}

	// Code and Reason should be preserved
	if clone.Code != original.Code || clone.Reason != original.Reason {
		t.Error("clone should preserve Code and Reason")
	}
}

func TestWithMetadata_AppendsToExisting(t *testing.T) {
	original := New(400, "ERR", "msg").
		WithMetadata(map[string]string{"a": "1"})

	clone := original.WithMetadata(map[string]string{"b": "2"})
	if clone.Metadata["a"] != "1" || clone.Metadata["b"] != "2" {
		t.Errorf("expected both a=1 and b=2, got %v", clone.Metadata)
	}
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	var _ error = (*Error)(nil)
}

func TestError_StdErrorsIs(t *testing.T) {
	// *Error should be comparable via errors.Is when same pointer
	original := New(404, "NOT_FOUND", "not found")
	if !stderrors.Is(original, original) {
		t.Error("errors.Is should return true for same pointer")
	}
}
