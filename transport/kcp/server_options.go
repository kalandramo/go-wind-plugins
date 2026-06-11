package kcp

import (
	"time"

	"github.com/tx7do/go-wind-plugins/encoding"
	"github.com/tx7do/go-wind-plugins/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////

type Option func(o *Server)

func WithAddress(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithCodec(c string) Option {
	return func(s *Server) {
		s.codec = encoding.GetCodec(c)
	}
}

func WithBlockCrypt(password, salt string) Option {
	return func(s *Server) {
		s.blockCryptPassword = password
		s.blockCryptSalt = salt
	}
}

func WithDataShards(dataShards int) Option {
	return func(s *Server) {
		s.dataShards = dataShards
	}
}

func WithParityShards(parityShards int) Option {
	return func(s *Server) {
		s.parityShards = parityShards
	}
}

func WithMessageMarshaler(m NetPacketMarshaler) Option {
	return func(s *Server) {
		s.netPacketMarshaler = m
	}
}

func WithMessageUnmarshaler(m NetPacketUnmarshaler) Option {
	return func(s *Server) {
		s.netPacketUnmarshaler = m
	}
}

func WithSocketConnectHandler(h SocketConnectHandler) Option {
	return func(s *Server) {
		s.socketConnectHandler = h
	}
}

func WithSocketRawDataHandler(h SocketRawDataHandler) Option {
	return func(s *Server) {
		if h != nil {
			s.socketRawDataHandler = h
		}
	}
}

// WithTracer 设置 OpenTelemetry tracer，启用消息级别的链路追踪。
func WithTracer(t trace.Tracer) Option {
	return func(s *Server) {
		s.tracer = t
	}
}

// WithTracerProvider 从全局 TracerProvider 创建并设置 tracer。
func WithTracerProvider() Option {
	return func(s *Server) {
		s.tracer = otel.GetTracerProvider().Tracer("go-wind/plugins/kcp")
	}
}

// WithMetrics 设置 metrics 实现，启用消息级别的指标采集。
func WithMetrics(m metrics.Metrics) Option {
	return func(s *Server) {
		s.m = m
	}
}
