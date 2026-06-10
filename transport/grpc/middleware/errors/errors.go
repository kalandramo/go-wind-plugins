// Package errors provides gRPC interceptors (server and client side) that
// translate between the framework's [errors.Error] type and gRPC
// [status.Status].
//
// On the server side, the interceptor inspects the error returned by the
// handler. If it wraps an [*errors.Error], the error's HTTP status code is
// mapped to a gRPC [codes.Code] and a proper [status.Status] is returned to the
// client.
//
// On the client side, the interceptor inspects the gRPC [status.Status] and
// converts it back into an [*errors.Error], so that client code can use
// [errors.FromError] uniformly regardless of the transport.
//
// Usage (server):
//
//	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
//	    grpcErrors.UnaryServerInterceptor(),
//	))
//
// Usage (client):
//
//	conn, _ := grpc.NewClient(addr,
//	    grpc.WithUnaryInterceptor(grpcErrors.UnaryClientInterceptor()),
//	)
package errors

import (
	"context"

	errs "github.com/tx7do/go-wind-plugins/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// HTTP ↔ gRPC code mapping
// ---------------------------------------------------------------------------

// httpToGRPC maps HTTP status codes to gRPC codes following the
// conventions used by grpc-gateway.
var httpToGRPC = map[int]codes.Code{
	200: codes.OK,
	400: codes.InvalidArgument,
	401: codes.Unauthenticated,
	403: codes.PermissionDenied,
	404: codes.NotFound,
	409: codes.Aborted,
	412: codes.FailedPrecondition,
	422: codes.InvalidArgument,
	429: codes.ResourceExhausted,
	500: codes.Internal,
	501: codes.Unimplemented,
	502: codes.Unavailable,
	503: codes.Unavailable,
	504: codes.DeadlineExceeded,
}

// grpcToHTTP maps gRPC codes to HTTP status codes following the conventions
// used by grpc-gateway.
var grpcToHTTP = map[codes.Code]int{
	codes.OK:                 200,
	codes.Canceled:           499,
	codes.Unknown:            500,
	codes.InvalidArgument:    400,
	codes.DeadlineExceeded:   504,
	codes.NotFound:           404,
	codes.AlreadyExists:      409,
	codes.PermissionDenied:   403,
	codes.ResourceExhausted:  429,
	codes.FailedPrecondition: 400,
	codes.Aborted:            409,
	codes.OutOfRange:         400,
	codes.Unimplemented:      501,
	codes.Internal:           500,
	codes.Unavailable:        503,
	codes.DataLoss:           500,
	codes.Unauthenticated:    401,
}

// HTTPToGRPC converts an HTTP status code to the closest gRPC code.
// Unmapped codes default to codes.Unknown.
func HTTPToGRPC(httpCode int) codes.Code {
	if c, ok := httpToGRPC[httpCode]; ok {
		return c
	}
	return codes.Unknown
}

// GRPCToHTTP converts a gRPC code to the closest HTTP status code.
// Unmapped codes default to 500.
func GRPCToHTTP(c codes.Code) int {
	if h, ok := grpcToHTTP[c]; ok {
		return h
	}
	return 500
}

// FromGRPCStatus converts a gRPC [status.Status] to an [*errors.Error].
//
// The gRPC code is mapped to an HTTP status code, and the status message is
// used as the error message. The gRPC code string is used as the Reason.
func FromGRPCStatus(st *status.Status) *errs.Error {
	return errs.New(
		GRPCToHTTP(st.Code()),
		st.Code().String(),
		st.Message(),
	)
}

// ToGRPCStatus converts an [*errors.Error] to a gRPC [status.Status].
func ToGRPCStatus(e *errs.Error) *status.Status {
	return status.New(HTTPToGRPC(e.Code), e.Message)
}

// ---------------------------------------------------------------------------
// Server-side interceptors
// ---------------------------------------------------------------------------

// UnaryServerInterceptor returns a [grpc.UnaryServerInterceptor] that converts
// [*errors.Error] values returned by handlers into gRPC status errors.
//
// If the handler returns a non-*errors.Error, it is passed through unchanged.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if e := errs.FromError(err); e != nil {
			return nil, ToGRPCStatus(e).Err()
		}
		return resp, err
	}
}

// StreamServerInterceptor returns a [grpc.StreamServerInterceptor] that
// converts [*errors.Error] values returned by handlers into gRPC status errors.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := handler(srv, ss)
		if err == nil {
			return nil
		}
		if e := errs.FromError(err); e != nil {
			return ToGRPCStatus(e).Err()
		}
		return err
	}
}

// ---------------------------------------------------------------------------
// Client-side interceptors
// ---------------------------------------------------------------------------

// UnaryClientInterceptor returns a [grpc.UnaryClientInterceptor] that converts
// gRPC status errors into [*errors.Error] values, so that client code can use
// [errs.FromError] uniformly regardless of transport.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}
		if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
			return FromGRPCStatus(st)
		}
		return err
	}
}

// StreamClientInterceptor returns a [grpc.StreamClientInterceptor] that
// converts gRPC status errors into [*errors.Error] values on stream creation.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err == nil {
			return stream, nil
		}
		if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
			return nil, FromGRPCStatus(st)
		}
		return stream, err
	}
}
