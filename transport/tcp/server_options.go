package tcp

import (
	"crypto/tls"
	"time"

	"github.com/tx7do/go-wind-plugins/encoding"
)

// Option 是 TCP 服务器的配置选项。
type Option func(*Server)

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

func WithTLSConfig(c *tls.Config) Option {
	return func(s *Server) {
		s.tlsConf = c
	}
}

func WithCodec(c string) Option {
	return func(s *Server) {
		s.codec = encoding.GetCodec(c)
	}
}

func WithChannelBufferSize(size int) Option {
	return func(_ *Server) {
		channelBufSize = size
	}
}

func WithReceiveBufferSize(size int) Option {
	return func(_ *Server) {
		recvBufferSize = size
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
