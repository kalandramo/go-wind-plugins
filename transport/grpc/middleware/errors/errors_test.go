package errors

import (
	"context"
	"fmt"
	"testing"

	errs "github.com/tx7do/go-wind-plugins/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Mapping functions
// ---------------------------------------------------------------------------

func TestHTTPToGRPC_KnownCodes(t *testing.T) {
	tests := []struct {
		http int
		grpc codes.Code
	}{
		{200, codes.OK},
		{400, codes.InvalidArgument},
		{401, codes.Unauthenticated},
		{403, codes.PermissionDenied},
		{404, codes.NotFound},
		{429, codes.ResourceExhausted},
		{500, codes.Internal},
		{501, codes.Unimplemented},
		{503, codes.Unavailable},
		{504, codes.DeadlineExceeded},
	}
	for _, tc := range tests {
		if got := HTTPToGRPC(tc.http); got != tc.grpc {
			t.Errorf("HTTPToGRPC(%d) = %v, want %v", tc.http, got, tc.grpc)
		}
	}
}

func TestHTTPToGRPC_UnknownCode(t *testing.T) {
	if got := HTTPToGRPC(599); got != codes.Unknown {
		t.Errorf("HTTPToGRPC(599) = %v, want Unknown", got)
	}
}

func TestGRPCToHTTP_KnownCodes(t *testing.T) {
	tests := []struct {
		grpc codes.Code
		http int
	}{
		{codes.OK, 200},
		{codes.InvalidArgument, 400},
		{codes.Unauthenticated, 401},
		{codes.PermissionDenied, 403},
		{codes.NotFound, 404},
		{codes.ResourceExhausted, 429},
		{codes.Internal, 500},
		{codes.Unimplemented, 501},
		{codes.Unavailable, 503},
		{codes.DeadlineExceeded, 504},
	}
	for _, tc := range tests {
		if got := GRPCToHTTP(tc.grpc); got != tc.http {
			t.Errorf("GRPCToHTTP(%v) = %d, want %d", tc.grpc, got, tc.http)
		}
	}
}

func TestGRPCToHTTP_UnknownCode(t *testing.T) {
	// codes.Unknown is in the map (→ 500), test a truly unmapped code
	if got := GRPCToHTTP(codes.Code(999)); got != 500 {
		t.Errorf("GRPCToHTTP(999) = %d, want 500", got)
	}
}

// ---------------------------------------------------------------------------
// FromGRPCStatus / ToGRPCStatus
// ---------------------------------------------------------------------------

func TestFromGRPCStatus(t *testing.T) {
	st := status.New(codes.NotFound, "user not found")
	e := FromGRPCStatus(st)

	if e.Code != 404 {
		t.Errorf("Code = %d, want 404", e.Code)
	}
	if e.Reason != codes.NotFound.String() {
		t.Errorf("Reason = %q", e.Reason)
	}
	if e.Message != "user not found" {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestToGRPCStatus(t *testing.T) {
	e := errs.New(401, "UNAUTHORIZED", "missing token")
	st := ToGRPCStatus(e)

	if st.Code() != codes.Unauthenticated {
		t.Errorf("Code = %v, want Unauthenticated", st.Code())
	}
	if st.Message() != "missing token" {
		t.Errorf("Message = %q", st.Message())
	}
}

func TestRoundTrip_GRPCStatusToErrorToGRPCStatus(t *testing.T) {
	original := status.New(codes.PermissionDenied, "forbidden")
	e := FromGRPCStatus(original)
	recovered := ToGRPCStatus(e)

	if recovered.Code() != original.Code() {
		t.Errorf("code mismatch: %v vs %v", recovered.Code(), original.Code())
	}
	if recovered.Message() != original.Message() {
		t.Errorf("message mismatch: %q vs %q", recovered.Message(), original.Message())
	}
}

// ---------------------------------------------------------------------------
// UnaryServerInterceptor
// ---------------------------------------------------------------------------

func TestUnaryServerInterceptor_ConvertsError(t *testing.T) {
	intc := UnaryServerInterceptor()

	handler := func(_ context.Context, _ any) (any, error) {
		return nil, errs.New(404, "USER_NOT_FOUND", "user not found")
	}

	_, err := intc(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Get"}, handler)
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
	if st.Message() != "user not found" {
		t.Errorf("message = %q", st.Message())
	}
}

func TestUnaryServerInterceptor_ConvertsWrappedError(t *testing.T) {
	intc := UnaryServerInterceptor()

	original := errs.New(403, "FORBIDDEN", "denied")
	handler := func(_ context.Context, _ any) (any, error) {
		return nil, fmt.Errorf("service: %w", original)
	}

	_, err := intc(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Do"}, handler)
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.PermissionDenied {
		t.Errorf("code = %v, want PermissionDenied", st.Code())
	}
}

func TestUnaryServerInterceptor_PassesThroughNonFrameworkError(t *testing.T) {
	intc := UnaryServerInterceptor()

	originalErr := status.Error(codes.DataLoss, "db corruption")
	handler := func(_ context.Context, _ any) (any, error) {
		return nil, originalErr
	}

	_, err := intc(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Get"}, handler)
	if err != originalErr {
		t.Error("non-framework gRPC error should pass through unchanged")
	}
}

func TestUnaryServerInterceptor_Success(t *testing.T) {
	intc := UnaryServerInterceptor()

	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	resp, err := intc(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Get"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

// ---------------------------------------------------------------------------
// StreamServerInterceptor
// ---------------------------------------------------------------------------

func TestStreamServerInterceptor_ConvertsError(t *testing.T) {
	intc := StreamServerInterceptor()

	handler := func(_ any, _ grpc.ServerStream) error {
		return errs.New(429, "RATE_LIMITED", "too many requests")
	}

	err := intc(nil, nil, &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}, handler)
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("code = %v, want ResourceExhausted", st.Code())
	}
}

func TestStreamServerInterceptor_Success(t *testing.T) {
	intc := StreamServerInterceptor()

	handler := func(_ any, _ grpc.ServerStream) error {
		return nil
	}

	err := intc(nil, nil, &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UnaryClientInterceptor
// ---------------------------------------------------------------------------

func TestUnaryClientInterceptor_ConvertsGRPCError(t *testing.T) {
	intc := UnaryClientInterceptor()

	invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		return status.Error(codes.NotFound, "user not found")
	}

	err := intc(context.Background(), "/svc/Get", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected error")
	}

	e := errs.FromError(err)
	if e == nil {
		t.Fatal("expected *errors.Error")
	}
	if e.Code != 404 {
		t.Errorf("HTTP code = %d, want 404", e.Code)
	}
	if e.Reason != codes.NotFound.String() {
		t.Errorf("Reason = %q", e.Reason)
	}
}

func TestUnaryClientInterceptor_PassesThroughSuccess(t *testing.T) {
	intc := UnaryClientInterceptor()

	invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		return nil
	}

	err := intc(context.Background(), "/svc/Get", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// StreamClientInterceptor
// ---------------------------------------------------------------------------

func TestStreamClientInterceptor_ConvertsError(t *testing.T) {
	intc := StreamClientInterceptor()

	streamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, status.Error(codes.Unavailable, "service down")
	}

	_, err := intc(context.Background(), nil, nil, "/svc/Stream", streamer)
	if err == nil {
		t.Fatal("expected error")
	}

	e := errs.FromError(err)
	if e == nil {
		t.Fatal("expected *errors.Error")
	}
	if e.Code != 503 {
		t.Errorf("HTTP code = %d, want 503", e.Code)
	}
}
