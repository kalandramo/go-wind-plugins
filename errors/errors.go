// Package errors defines a protocol-agnostic error type for the go-wind
// framework.
//
// Unlike frameworks that generate errors from .proto files (requiring the
// protobuf toolchain), this package defines errors in pure Go code. This makes
// the error type usable across all transport protocols (HTTP, gRPC, GraphQL,
// Thrift, etc.) without forcing a specific IDL.
//
// The Error type uses HTTP status codes as its semantic anchor. Each transport
// layer then maps the code to its own representation:
//
//   - HTTP: the code is used directly as the HTTP status code.
//   - gRPC: the code is mapped to a [codes.Code] via the mapping table in
//     transport/grpc/middleware/errors.
//
// Usage:
//
//	// Define domain errors
//	var ErrUserNotFound = errors.New(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
//
//	// Use in business logic
//	func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
//	    user, err := s.repo.Find(id)
//	    if err != nil {
//	        return nil, ErrUserNotFound
//	    }
//	    return user, nil
//	}
package errors

import (
	"errors"
	"fmt"
)

// Error is the protocol-agnostic error type shared across all transport layers.
//
// Code is the primary semantic field: it is an HTTP status code (e.g. 404,
// 400, 500) that serves as the universal anchor. Each transport layer maps
// this code to its own representation.
type Error struct {
	// Code is the HTTP status code used as the semantic anchor.
	Code int

	// Reason is a machine-readable, stable identifier for the error.
	// Examples: "USER_NOT_FOUND", "INVALID_ARGUMENT".
	Reason string

	// Message is a human-readable description of the error.
	Message string

	// Metadata contains optional key-value pairs for additional context
	// (e.g. locale-specific messages, trace IDs, field details).
	Metadata map[string]string
}

// Error implements the [error] interface.
func (e *Error) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("%s: %s", e.Reason, e.Message)
	}
	return e.Message
}

// WithMetadata returns a shallow copy of the error with the given metadata
// merged in. The original error is not modified.
func (e *Error) WithMetadata(kv map[string]string) *Error {
	clone := &Error{
		Code:    e.Code,
		Reason:  e.Reason,
		Message: e.Message,
	}
	if len(e.Metadata) > 0 {
		clone.Metadata = make(map[string]string, len(e.Metadata)+len(kv))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}
	for k, v := range kv {
		if clone.Metadata == nil {
			clone.Metadata = make(map[string]string)
		}
		clone.Metadata[k] = v
	}
	return clone
}

// New creates a new [*Error] with the given code, reason, and message.
func New(code int, reason, message string) *Error {
	return &Error{
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// Newf creates a new [*Error] with a formatted message.
func Newf(code int, reason, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Reason:  reason,
		Message: fmt.Sprintf(format, args...),
	}
}

// FromError extracts an [*Error] from a standard error.
//
// It uses [errors.As] so that errors wrapped with %w are handled correctly:
//
//	wrapped := fmt.Errorf("service layer: %w", ErrUserNotFound)
//	e := errors.FromError(wrapped) // e == ErrUserNotFound
//
// If err is nil or does not contain an [*Error], FromError returns nil.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// Code returns the HTTP status code associated with err.
//
// If err wraps an [*Error], its Code field is returned.
// Otherwise [StatusInternalServerError] (500) is returned as the safe default.
func Code(err error) int {
	if err == nil {
		return 200
	}
	if e := FromError(err); e != nil {
		return e.Code
	}
	return 500
}

// HTTP status code constants for convenience. These mirror the constants in
// [net/http] but are defined here so that the errors package has no external
// dependencies.
const (
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusPreconditionFailed  = 412
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
	StatusNotImplemented      = 501
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusGatewayTimeout      = 504
)
