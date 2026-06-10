package thrift

import (
	"crypto/tls"

	"github.com/apache/thrift/lib/go/thrift"
)

// Option 是 Thrift 服务器的配置选项。
type Option func(*Server)

// WithTLSConfig 设置 TLS 配置，启用加密传输。
func WithTLSConfig(c *tls.Config) Option {
	return func(s *Server) { s.tlsConfig = c }
}

// WithProcessor 设置 Thrift 请求处理器（必需）。
// processor 由 thrift IDL 编译器生成的代码创建，例如：
//
//	processor := api.NewMyServiceProcessor(handler)
func WithProcessor(processor thrift.TProcessor) Option {
	return func(s *Server) { s.processor = processor }
}

// WithProtocol 设置 Thrift 协议类型。
// 支持的值："binary"（默认）, "compact", "json"。
func WithProtocol(protocol string) Option {
	return func(s *Server) { s.protocol = protocol }
}

// WithTransportConfig 配置传输层参数。
//   - buffered: 启用缓冲传输
//   - framed: 启用帧传输（非阻塞服务必需）
//   - bufferSize: 缓冲区大小（字节），默认 8192
func WithTransportConfig(buffered, framed bool, bufferSize int) Option {
	return func(s *Server) {
		s.buffered = buffered
		s.framed = framed
		s.bufferSize = bufferSize
	}
}
